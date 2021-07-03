package parser

import (
	"fmt"
	"nala/ast"
	"nala/lexer"
	"nala/token"
	"strconv"
)

const (
	_ int = iota // this gives the constants incrementing numbers as values
	LOWEST
	EQUALS      // ==
	LESSGREATER // < or >
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x or !x
	CALL        // funcName(var, y)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MODULO:   PRODUCT,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression // parameter is the lhs of the expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// this sets both curToken and peekToken
	p.nextToken()
	p.nextToken()

	// associate functions with TokenTypes
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	prog.Statements = []ast.Statement{}

	// iterate over tokens till EOF is seen
	for !p.curTokenIs(token.EOF) {
		stmt := p.ParseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
		p.nextToken()
	}

	return prog
}

func (p *Parser) ParseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer untrace(trace("parseLetStatement"))
	stmt := &ast.LetStatement{Token: p.curToken}

	// this makes sure the next token is an IDENT token
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// TODO: Expressions will be handled here. We skip over it for now
	// there is a bug here. If u skip the terminating ';' it enters an infinite loop
	// for !p.curTokenIs(token.SEMICOLON) {
	// 	if !p.curTokenIs(token.EOF) {
	// 		p.nextToken()
	// 	} else {
	// 		break
	// 	}
	// }

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer untrace(trace("parseReturnStatement"))
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// TODO: Expressions will be handled here. We skip over it for now
	// there is a bug here. If u skip the terminating ';' it enters an infinite loop
	for !p.curTokenIs(token.SEMICOLON) {
		if !p.curTokenIs(token.EOF) {
			p.nextToken()
		} else {
			break
		}
	}

	if !p.expectCur(token.SEMICOLON) {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError()
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	defer untrace(trace("parseIdentifier"))
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}

	val, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.integerParseError()
		return nil
	}

	lit.Value = val
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) curPrecedence() int {
	return findPrecedence(p.curToken.Type)
}

func (p *Parser) peekPrecedence() int {
	return findPrecedence(p.peekToken.Type)
}

func findPrecedence(tokType token.TokenType) int {
	if p, ok := precedences[tokType]; ok {
		return p
	}
	return LOWEST
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

func (p *Parser) curTokenIs(expectedType token.TokenType) bool {
	return expectedType == p.curToken.Type
}

func (p *Parser) peekTokenIs(expectedType token.TokenType) bool {
	return expectedType == p.peekToken.Type
}

func (p *Parser) Errors() []string { return p.errors }

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

func (p *Parser) noPrefixParseFnError() {
	err := fmt.Sprintf("no prefix parse function found for %s",
		p.curToken.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) noInfixParseFnError() {
	err := fmt.Sprintf("no infix parse function found for %s",
		p.curToken.Type)
	p.errors = append(p.errors, err)
}
