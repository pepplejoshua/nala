package lexer

import (
	"nala/token"
	"testing"
)

type ExpectedToken struct {
	expectedType    token.TokenType
	expectedLiteral string
}

func TestNextToken(t *testing.T) {
	input := `=+(){},;`

	// generates an array of expected tokens from that initializer list
	tests := []ExpectedToken{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
	}

	// makes a new Lexer?
	l := New(input)

	for indx, expTok := range tests {
		tok := l.NextToken()

		if tok.Type != expTok.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", indx, expTok.expectedType, tok.Type)
		}

		if tok.Literal != expTok.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", indx, expTok.expectedType, tok.Type)
		}
	}
}
