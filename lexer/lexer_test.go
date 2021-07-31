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
	// input := `=+(){},;`
	input := `let five = 5;
	let ten = 10;
	
	let add = fn(x, y) {
		x + y;	
	};
	
	let result = add(five, ten);
	!-/*%5;
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 10;
	10 != 9;
	"foobar";
	"foo bar"
	[1, 2, 10 > 5];
	{ "foo" : "bar" }
	macro(a, b) { a + b };
	`

	// generates an array of expected tokens from that initializer list
	tests := []ExpectedToken{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},

		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},

		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.MODULO, "%"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},

		{token.STRING, "foobar"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foo bar"},

		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},

		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},

		{token.MACRO, "macro"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},

		{token.EOF, ""},
	}

	// makes a new Lexer?
	l := New(input)

	for indx, expTok := range tests {
		tok := l.NextToken()

		if tok.Type != expTok.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", indx, expTok.expectedType, tok.Type)
		}

		if tok.Literal != expTok.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", indx, expTok.expectedLiteral, tok.Literal)
		}
	}
}

func TestEllispNextToken(t *testing.T) {
	input := `
	(let five 5)
	(let ten 10)

	(let (add x y) (+ x y))
	(fn (x y) (+ x y))

	(== 10 10)
	(!= 10 10)
	(< 5 (> 10 5))
	(!true)
	(- 5)

	-5
	"joshua pepple"
	true
	false
	'()
	[1, 2, (< 3 2), 4]
	[1 2 3 4]
	{"1": 1, "2": 2}
	{"1": 1 true: 2}
	(cons 1 '())
	(list 1 2 3)
	'(1 2 3 4)

	(let res (add five ten))

	(if (< 5 10)
		true
		false)
	`

	tests := []ExpectedToken{
		{token.LPAREN, "("},
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.INT, "5"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.INT, "10"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.LET, "let"},
		{token.LPAREN, "("},
		{token.IDENT, "add"},
		{token.IDENT, "x"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LPAREN, "("},
		{token.PLUS, "+"},
		{token.IDENT, "x"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LPAREN, "("},
		{token.PLUS, "+"},
		{token.IDENT, "x"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.INT, "10"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.NOT_EQ, "!="},
		{token.INT, "10"},
		{token.INT, "10"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.LT, "<"},
		{token.INT, "5"},
		{token.LPAREN, "("},
		{token.GT, ">"},
		{token.INT, "10"},
		{token.INT, "5"},
		{token.RPAREN, ")"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.BANG, "!"},
		{token.TRUE, "true"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.MINUS, "-"},
		{token.INT, "5"},
		{token.RPAREN, ")"},

		{token.MINUS, "-"},
		{token.INT, "5"},

		{token.STRING, "joshua pepple"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},

		{token.APOSTROPHE, "'"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},

		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.LPAREN, "("},
		{token.LT, "<"},
		{token.INT, "3"},
		{token.INT, "2"},
		{token.RPAREN, ")"},
		{token.COMMA, ","},
		{token.INT, "4"},
		{token.RBRACKET, "]"},

		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.INT, "2"},
		{token.INT, "3"},
		{token.INT, "4"},
		{token.RBRACKET, "]"},

		{token.LBRACE, "{"},
		{token.STRING, "1"},
		{token.COLON, ":"},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.STRING, "2"},
		{token.COLON, ":"},
		{token.INT, "2"},
		{token.RBRACE, "}"},

		{token.LBRACE, "{"},
		{token.STRING, "1"},
		{token.COLON, ":"},
		{token.INT, "1"},
		{token.TRUE, "true"},
		{token.COLON, ":"},
		{token.INT, "2"},
		{token.RBRACE, "}"},

		{token.LPAREN, "("},
		{token.CONS, "cons"},
		{token.INT, "1"},
		{token.APOSTROPHE, "'"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.LIST, "list"},
		{token.INT, "1"},
		{token.INT, "2"},
		{token.INT, "3"},
		{token.RPAREN, ")"},

		{token.APOSTROPHE, "'"},
		{token.LPAREN, "("},
		{token.INT, "1"},
		{token.INT, "2"},
		{token.INT, "3"},
		{token.INT, "4"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.LET, "let"},
		{token.IDENT, "res"},
		{token.LPAREN, "("},
		{token.IDENT, "add"},
		{token.IDENT, "five"},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.RPAREN, ")"},

		{token.LPAREN, "("},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.LT, "<"},
		{token.INT, "5"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.RPAREN, ")"},
	}

	l := New(input)

	for indx, expTok := range tests {
		tok := l.NextToken()

		if tok.Type != expTok.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", indx, expTok.expectedType, tok.Type)
		}

		if tok.Literal != expTok.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", indx, expTok.expectedLiteral, tok.Literal)
		}
	}

}
