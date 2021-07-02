package parser

import (
	"fmt"
	"nala/ast"
	"nala/lexer"
	"nala/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// this sets both curToken and peekToken
	p.nextToken()
	p.nextToken()

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
		return nil
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// this makes sure the next token is an IDENT token
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

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

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
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