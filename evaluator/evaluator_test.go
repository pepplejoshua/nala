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

type GenericTest struct {
	input    string
	expected interface{}
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
		t.Errorf("object has wrong value. got=%d,  want=%d", res.Value, expected)
		return false
	}
	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []BooleanTest{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"true != true", false},
		{"false == false", true},
		{"false != false", false},
		{"true == false", false},
		{"true != false", true},
		{"(1 < 2) == true", true},
		{"1 != 2 == true", true},
		{"false != (1 == 1)", true},
		{"true != (2 != 1)", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, evalObj object.Object, expected bool) bool {
	res, ok := evalObj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", evalObj, evalObj)
		return false
	}
	if res.Value != expected {
		t.Errorf("object has wrong value. got=%v,  want=%v", res.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, evalObj object.Object) bool {
	if evalObj != NIL {
		t.Errorf("object is not NIL. got=%T (%+v)", evalObj, evalObj)
		return false
	}
	return true
}

func testEvalLiteral(t *testing.T, evalObj object.Object, expected interface{}) bool {
	switch expected := expected.(type) {
	case int:
		return testIntegerObject(t, evalObj, int64(expected))
	case bool:
		return testBooleanObject(t, evalObj, expected)
	default:
		return testNullObject(t, evalObj)
	}
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

func TestEvalIfElseExpressions(t *testing.T) {
	tests := []GenericTest{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 0 }", 0},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (true) { false }", false},
		{"if (false) { false } else { true }", true},
		{"if (1) { false } else { true }", false},
		{"if (0) { false } else { true }", true},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)

		testEvalLiteral(t, evald, tt.expected)
	}
}

func TestEvalReturnStatements(t *testing.T) {
	tests := []GenericTest{
		{"return 10;", 10},
		{"return 11; 4", 11},
		{"return 2*15; 9;", 30},
		{"9; return 2*16; 9", 32},
		{
			`if (10 > 1) {
				if (true) {
					return 200;		
				}
			}
			return 1;`,
			200,
		},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)

		testEvalLiteral(t, evald, tt.expected)
	}
}

func TestEvalErrorHandling(t *testing.T) {
	tests := []GenericTest{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true;",
			"unknown operator: -BOOLEAN",
		},
		{
			"true - false;",
			"unknown operator: BOOLEAN - BOOLEAN",
		},
		{
			"5; true * false; 5",
			"unknown operator: BOOLEAN * BOOLEAN",
		},
		{
			"if (10 > 1) { true / false; }",
			"unknown operator: BOOLEAN / BOOLEAN",
		},
		{
			`if (10 > 1) {
				if (true) {
					return true % false;		
				}
			}
			return 1;`,
			"unknown operator: BOOLEAN % BOOLEAN",
		},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)

		err, ok := evald.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evald, evald)
			continue
		}

		if err.Message != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, err.Message)
		}
	}
}
