package lispparser

import (
	"fmt"
	"nala/ast"
	"nala/lexer"
	"nala/token"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
)

// statements
// let
// return

// prefixes:
// lparen
// int
// string
// true
// false
// array [lbracket]
// hmap {lbrace}
// identifier

// coerced into prefix forms from lparen:
// +, -, /, *, !, !=, ==, <, >, []
// if
// let
// fn (literals)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// this sets both curToken and peekToken
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.LPAREN, p.parseParenthesesExpression)
	p.registerPrefix(token.IDENT, p.parseLiteral)
	p.registerPrefix(token.INT, p.parseLiteral)
	p.registerPrefix(token.STRING, p.parseLiteral)
	p.registerPrefix(token.TRUE, p.parseLiteral)
	p.registerPrefix(token.FALSE, p.parseLiteral)

	// explicit coerced prefixes
	p.registerPrefix(token.BANG, p.parseNegateBooleanExpression)
	p.registerPrefix(token.MINUS, p.parseNaryExpression)

	p.registerPrefix(token.PLUS, p.parseBinaryExpression)
	p.registerPrefix(token.ASTERISK, p.parseBinaryExpression)
	p.registerPrefix(token.SLASH, p.parseBinaryExpression)
	p.registerPrefix(token.MODULO, p.parseBinaryExpression)
	p.registerPrefix(token.EQ, p.parseBinaryExpression)
	p.registerPrefix(token.NOT_EQ, p.parseBinaryExpression)
	p.registerPrefix(token.LT, p.parseBinaryExpression)
	p.registerPrefix(token.GT, p.parseBinaryExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	// p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	prog.Statements = []ast.Statement{}

	// iterate over tokens till EOF is seen
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
		p.nextToken()
	}

	return prog
}

func (p *Parser) parseStatement() ast.Statement {
	// decide where to route let and return statements.
	// as they can both be easily returned from there.
	switch p.curToken.Type {
	case token.LPAREN: // could be a let or return statement so let's handle early
		return p.parseParenthesesStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseParenthesesExpression() ast.Expression {
	p.nextToken()
	return p.parseExpression()
}

func (p *Parser) parseParenthesesStatement() ast.Statement {
	if p.peekTokenIs(token.LET) {
		p.nextToken() // advance to the let token
		return p.parseLetStatement()
	} else if p.peekTokenIs(token.RETURN) {
		p.nextToken()
		return p.parseReturnStatement()
	} else {
		tok := p.peekToken
		expr := p.parseParenthesesExpression()
		return &ast.ExpressionStatement{
			Token:      tok,
			Expression: expr,
		}
	}

	// else if p.peekTokenIs(token.IDENT) {
	// 	p.nextToken()
	// 	return p.parseCallExpression()

}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{
		Token: p.curToken,
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	p.nextToken()
	stmt.Value = p.parseExpression()

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression()

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return stmt
}

func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expr.Condition = p.parseExpression()
	if !p.expectPeek(token.COLON) { // start of consequence
		return nil
	}
	expr.Consequence = p.parseBlockStatement(token.COMMA, token.RPAREN)

	if p.curTokenIs(token.COMMA) { // start of else block
		expr.Alternative = p.parseBlockStatement(token.RPAREN, token.RPAREN)
	}

	if !p.expectCur(token.RPAREN) {
		return nil
	}
	p.nextToken()
	return expr
}

func (p *Parser) parseBlockStatement(endToken token.TokenType, fallbackEnd token.TokenType) *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(endToken) && !p.curTokenIs(fallbackEnd) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression()

	return stmt
}

func (p *Parser) parseBinaryExpression() ast.Expression {
	sign := p.curToken

	p.nextToken()
	left := p.parseExpression()
	p.nextToken()
	right := p.parseExpression()

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return &ast.InfixExpression{
		Token:    sign,
		Left:     left,
		Operator: sign.Literal,
		Right:    right,
	}
}

func (p *Parser) parseNegateBooleanExpression() ast.Expression {
	sign := p.curToken
	p.nextToken()
	left := p.parseExpression()

	expr := &ast.PrefixExpression{
		Token:    sign,
		Operator: sign.Literal,
		Right:    left,
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return expr
}
func (p *Parser) parseNaryExpression() ast.Expression {
	sign := p.curToken
	p.nextToken()

	left := p.parseExpression()
	var expr ast.Expression

	if !p.peekTokenIs(token.RPAREN) {
		// then its a binary expression
		p.nextToken()
		expr = &ast.InfixExpression{
			Token:    sign,
			Left:     left,
			Operator: sign.Literal,
			Right:    p.parseExpression(),
		}
	} else {
		expr = &ast.PrefixExpression{
			Token:    sign,
			Operator: sign.Literal,
			Right:    left,
		}
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return expr
}

func (p *Parser) parseExpression() ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError()
		return nil
	}
	leftExp := prefix()

	return leftExp
}

func (p *Parser) parseLiteral() ast.Expression {
	switch p.curToken.Type {
	case token.INT:
		lit := &ast.IntegerLiteral{Token: p.curToken}

		val, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
		if err != nil {
			p.integerParseError()
			return nil
		}
		lit.Value = val
		return lit
	case token.TRUE, token.FALSE:
		return &ast.Boolean{
			Token: p.curToken,
			Value: p.curTokenIs(token.TRUE),
		}
	case token.STRING:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.IDENT:
		return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	default:
		p.nonLiteralError()
		return nil
	}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fn.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.COLON) {
		return nil
	}

	fn.Body = p.parseBlockStatement(token.RPAREN, token.RPAREN)
	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	ids := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return ids
	}

	p.nextToken()

	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	ids = append(ids, id)

	fmt.Println(id.String())
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		ids = append(ids, id)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return ids
}

func (p *Parser) curTokenIs(expectedType token.TokenType) bool {
	return expectedType == p.curToken.Type
}

func (p *Parser) peekTokenIs(expectedType token.TokenType) bool {
	return expectedType == p.peekToken.Type
}

func (p *Parser) expectPeek(expectedType token.TokenType) bool {
	if p.peekTokenIs(expectedType) {
		p.nextToken()
		return true
	} else {
		p.peekError(expectedType)
		return false
	}
}

func (p *Parser) expectCur(expectedType token.TokenType) bool {
	if p.curTokenIs(expectedType) {
		return true
	} else {
		p.curError(expectedType)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	err := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) curError(t token.TokenType) {
	err := fmt.Sprintf("expected current token to be %s, got %s instead",
		t, p.curToken.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) integerParseError() {
	err := fmt.Sprintf("could not parse %q as integer",
		p.curToken.Literal)
	p.errors = append(p.errors, err)
}

func (p *Parser) nonLiteralError() {
	err := fmt.Sprintf("could not parse %q as a literal",
		p.curToken.Literal)
	p.errors = append(p.errors, err)
}

func (p *Parser) noPrefixParseFnError() {
	err := fmt.Sprintf("no prefix parse function found for %s",
		p.curToken.Type)
	p.errors = append(p.errors, err)
}
