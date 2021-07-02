package ast

import (
	"nala/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "joshua"},
					Value: "joshua",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "otherVar"},
					Value: "otherVar",
				},
			},
		},
	}

	if program.String() != "let joshua = otherVar;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
