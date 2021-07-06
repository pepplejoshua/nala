package ast

import (
	"reflect"
	"testing"
)

type GenericTest struct {
	input    Node
	expected Node
}

func TestModify(t *testing.T) {
	one := func() Expression { return &IntegerLiteral{Value: 1} }
	four := func() Expression { return &IntegerLiteral{Value: 4} }

	turnOneIntoFour := func(node Node) Node {
		integer, ok := node.(*IntegerLiteral)
		if !ok {
			return node
		}

		if integer.Value != 1 {
			return node
		}

		integer.Value = 4
		return integer
	}

	tests := []GenericTest{
		{
			one(),
			four(),
		},
		{
			&Program{
				Statements: []Statement{
					&ExpressionStatement{Expression: one()},
				},
			},
			&Program{
				Statements: []Statement{
					&ExpressionStatement{Expression: four()},
				},
			},
		},
		{
			&InfixExpression{Left: one(), Operator: "+", Right: one()},
			&InfixExpression{Left: four(), Operator: "+", Right: four()},
		},
		{
			&InfixExpression{Left: four(), Operator: "+", Right: one()},
			&InfixExpression{Left: four(), Operator: "+", Right: four()},
		},
		{
			&PrefixExpression{Operator: "-", Right: one()},
			&PrefixExpression{Operator: "-", Right: four()},
		},
		{
			&IndexExpression{Left: one(), Index: one()},
			&IndexExpression{Left: four(), Index: four()},
		},
		{
			&IfExpression{
				Condition: one(),
				Consequence: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: one()},
					},
				},
				Alternative: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: one()},
					},
				},
			},
			&IfExpression{
				Condition: four(),
				Consequence: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: four()},
					},
				},
				Alternative: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: four()},
					},
				},
			},
		},
		{
			&ReturnStatement{ReturnValue: one()},
			&ReturnStatement{ReturnValue: four()},
		},
		{
			&LetStatement{Value: one()},
			&LetStatement{Value: four()},
		},
		{
			&FunctionLiteral{
				Parameters: []*Identifier{},
				Body: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: one()},
					},
				},
			},
			&FunctionLiteral{
				Parameters: []*Identifier{},
				Body: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: four()},
					},
				},
			},
		},
		{
			&ArrayLiteral{
				Elements: []Expression{one(), one(), four(), one()},
			},
			&ArrayLiteral{
				Elements: []Expression{four(), four(), four(), four()},
			},
		},
	}

	for _, tt := range tests {
		mod := Modify(tt.input, turnOneIntoFour)

		equal := reflect.DeepEqual(mod, tt.expected)
		if !equal {
			t.Errorf("not equal. got=%#v, want=%#v", mod, tt.expected)
		}
	}

	hashLit := &HashLiteral{
		Pairs: ExpressionPairs{
			one(): one(),
			one(): four(),
		},
	}

	Modify(hashLit, turnOneIntoFour)

	for k, v := range hashLit.Pairs {
		k, _ := k.(*IntegerLiteral)
		if k.Value != 4 {
			t.Errorf("value is not %d, got=%d", 2, k.Value)
		}

		v, _ := v.(*IntegerLiteral)
		if v.Value != 4 {
			t.Errorf("value is not %d, got=%d", 2, v.Value)
		}
	}
}
