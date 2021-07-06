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

type FunctionTest struct {
	input        string
	paramLen     int
	params       []string
	expectedBody string
}

type HashMapTests map[object.HashKey]int64

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
	env := object.NewEnvironment()
	return Eval(prog, env)
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
		{`"joshua" == "joshua"`, true},
		{`"pepple" != "iwarilama"`, true},
		{`"joshua" != "joshua"`, false},
		{`"pepple" == "iwarilama"`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		t.Log(tt.input)
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
	case string:
		return testStringObject(t, evalObj, expected)
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
		{
			"pepple;",
			"identifier not found: pepple",
		},
		{
			"tamunoiwarilama",
			"identifier not found: tamunoiwarilama",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name" : "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
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

func TestEvalLetStatements(t *testing.T) {
	tests := []GenericTest{
		{"let joshua = 5; joshua;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c", 15},
		{"let joshua = false; joshua;", false},
		{"let a = 1 < 2; a;", true},
		{"let a = true; let b = a; b;", true},
		{"let a = false; let b = a; b;", false},
		{"let a = true; let b = a; let c = a == b != false; c", true},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)
		testEvalLiteral(t, evald, tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	tests := []FunctionTest{
		{
			input:        "fn(x) { x + 2; };",
			paramLen:     1,
			params:       []string{"x"},
			expectedBody: "    (x + 2)",
		},
		{
			input:        "fn(x, y) { x + y; };",
			paramLen:     2,
			params:       []string{"x", "y"},
			expectedBody: "    (x + y)",
		},
		{
			input:        "fn(x, y, z) { x + y; z };",
			paramLen:     3,
			params:       []string{"x", "y", "z"},
			expectedBody: "    (x + y)    z",
		},
		{
			input:        "fn(x, y, z) { let z = x + y; return z; y};",
			paramLen:     3,
			params:       []string{"x", "y", "z"},
			expectedBody: "    let z = (x + y);    return z;    y",
		},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)

		fn, ok := evald.(*object.Function)
		if !ok {
			t.Fatalf("object is not a Function. got=%T (%+v)", evald, evald)
		}

		if len(fn.Parameters) != tt.paramLen {
			t.Fatalf("function has wrong parameters. Parameters=%+v", fn.Parameters)
		}

		for i, p := range fn.Parameters {
			if p.String() != tt.params[i] {
				t.Fatalf("parameter is not '%q'. got=%q", tt.params[i], p.String())
			}
		}

		if fn.Body.String() != tt.expectedBody {
			t.Fatalf("body is not %q. got=%q", tt.expectedBody, fn.Body.String())
		}
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []GenericTest{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { x; }; identity(false);", false},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let lessThan = fn(x, y) { x < y }; lessThan(4, 5)", true},
		{"let lessThan = fn(x, y) { x < y }; lessThan(10, 5)", false},
		{"let equal? = fn(x, y) { !(x != y) }; equal?(2, 3)", false},
		{"fn(x) { 3 * x }(5)", 15},
		{"fn(x, y) { y * x }(5, 3)", 15},
	}

	for _, tt := range tests {
		testEvalLiteral(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let adderTemplate* = fn(x) {
		fn(y) { x + y }
	};
	let addTwo = adderTemplate*(2);
	addTwo(8)
	`

	testIntegerObject(t, testEval(input), 10)
}

func TestEvalStringLiterals(t *testing.T) {
	tests := []GenericTest{
		{`"hello world"`, "hello world"},
		{`"joshua pepple"`, "joshua pepple"},
		{`"working on compiler"`, "working on compiler"},
		{`"a"`, "a"},
		{`""`, ""},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)
		testStringObject(t, evald, tt.expected)
	}
}

func testStringObject(t *testing.T, evalObj object.Object, expected interface{}) bool {
	str, ok := evalObj.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evalObj, evalObj)
		return false
	}

	if str.Value != expected {
		t.Errorf("String has wrong value. expected=%q, got=%q", expected, str.Value)
		return false
	}
	return true
}

func TestStringConcatenationExpressions(t *testing.T) {
	tests := []GenericTest{
		{`"Hello" + " " + "joshua"`, "Hello joshua"},
		{`"Joshua" + " " + "Pepple"`, "Joshua Pepple"},
		{`"1"+"2"`, "12"},
		{`""+""`, ""},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)
		testStringObject(t, evald, tt.expected)
	}
}

func TestBuiltInFunctions(t *testing.T) {
	tests := []GenericTest{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` is not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		evald := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evald, int64(expected))
		case string:
			err, ok := evald.(*object.Error)
			if !ok {
				t.Errorf("object is not an Error object. got=%T (%+v)", evald, evald)
			}
			if err.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, err.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evald := testEval(input)
	res, ok := evald.(*object.Array)
	if !ok {
		t.Errorf("object is not an Array object. got=%T (%+v)", evald, evald)
	}

	if len(res.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(res.Elements))
	}

	testIntegerObject(t, res.Elements[0], 1)
	testIntegerObject(t, res.Elements[1], 4)
	testIntegerObject(t, res.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []GenericTest{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"let i = 0; [1, 2, 3][i]", 1},
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1+1]", 3},
		{"let arr = [1, 2, 3, 4]; arr[3]", 4},
		{"let arr = [1, 2, 3, 4]; arr[3] + arr[1] + arr[0]", 7},
		{"let arr = [1, 2, 3, 4]; let i = arr[0]; arr[i]", 2},
		{"[1, 2, 3][3]", nil},
		{"[1, 2, 3][-1]", nil},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		Int, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(Int))
		} else {
			testNullObject(t, eval)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `
	let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	eval := testEval(input)
	res, ok := eval.(*object.HashMap)
	if !ok {
		t.Errorf("Eval didn't return Hash. got=%T (%+v)", eval, eval)
	}

	expected := HashMapTests{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(res.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong number of pairs. got=%d", len(res.Pairs))
	}

	for eKey, eVal := range expected {
		p, ok := res.Pairs[eKey]

		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, p.Value, eVal)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []GenericTest{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5, true: 4, 1: 2, 3: 4}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		Int, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(Int))
		} else {
			testNullObject(t, eval)
		}
	}
}

func TestQuotes(t *testing.T) {
	tests := []GenericTest{
		{`quote(5);`, `5`},
		{`quote(2+2);`, `(2 + 2)`},
		{`quote(foobar);`, `foobar`},
		{`quote(foobar + barfoo);`, `(foobar + barfoo)`},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testQuote(t, eval, tt.expected)
	}
}

func TestQuoteUnquotes(t *testing.T) {
	tests := []GenericTest{
		{`quote(unquote(4))`, `4`},
		{`quote(8 + unquote(4 + 4))`, `(8 + 8)`},
		{`quote(unquote(true) == 1 < 2)`, `(true == (1 < 2))`},
		{`quote(unquote(4 + 4) % 8)`, `(8 % 8)`},
		{`let f = 8; quote(f)`, `f`},
		{`let f = 8; quote(unquote(f))`, `8`},
		{
			`quote(unquote(quote(4 + 4)))`,
			`(4 + 4)`,
		},
		{
			`let qIE = quote(4 + 4);
			quote(unquote(4 + 4) + unquote(qIE))`,
			`(8 + (4 + 4))`,
		},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testQuote(t, eval, tt.expected)
	}
}

func testQuote(t *testing.T, res object.Object, expStr interface{}) bool {
	quote, ok := res.(*object.Quote)
	if !ok {
		t.Fatalf("expected *object.Quote. got=%T (%+v)", res, res)
		return false
	}

	if quote.CodeNode == nil {
		t.Fatalf("quote.Node is nil")
		return false
	}

	if quote.CodeNode.String() != expStr {
		t.Errorf("not equal. got=%q, want=%q", quote.CodeNode.String(), expStr)
		return false
	}
	return true
}
