package evaluator

import (
	"nala/lexer"
	"nala/object"
	"nala/parser"
	"testing"
)

type IntegerTest struct {
	input    string
	expected int64
}

type BooleanTest struct {
	input    string
	expected bool
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []IntegerTest{
		{"5", 5},
		{"2000", 2000},
		{"-5", -5},
		{"-24", -24},
		{"5 + 5 + 5 + 5 + 5 - 10", 15},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"5 * 5 % 6", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	return Eval(prog)
}

func testIntegerObject(t *testing.T, evalObj object.Object, expected int64) bool {
	res, ok := evalObj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", evalObj, evalObj)
		return false
	}
	if res.Value != expected {
		t.Errorf("object has wrong value. got=%d,  want=%d", expected, res.Value)
		return false
	}
	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []BooleanTest{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, evalObj object.Object, expected bool) bool {
	res, ok := evalObj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", evalObj, evalObj)
		return false
	}
	if res.Value != expected {
		t.Errorf("object has wrong value. got=%v,  want=%v", expected, res.Value)
		return false
	}
	return true
}

func TestBangOperator(t *testing.T) {
	tests := []BooleanTest{
		{"!false", true},
		{"!true", false},
		{"!4", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}
