package vm

import (
	"fmt"
	"nala/ast"
	"nala/compiler"
	"nala/lexer"
	"nala/object"
	"nala/parser"
	"testing"
)

type vmTest struct {
	input    string
	expected interface{}
}

func parse(in string) ast.Node {
	return parser.New(lexer.New(in)).ParseProgram()
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTest{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"5 + 2 * 10", 25},
	}

	runVmTests(t, tests)
}

func testIntegerObject(exp int64, act object.Object) error {
	res, ok := act.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", act, act)
	}

	if res.Value != exp {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", res.Value, exp)
	}
	return nil
}

func testBooleanObject(exp bool, act object.Object) error {
	res, ok := act.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", act, act)
	}

	if res.Value != exp {
		return fmt.Errorf("object has wrong value. got=%v, want=%v", res.Value, exp)
	}
	return nil
}

func runVmTests(t *testing.T, tests []vmTest) {
	t.Helper()

	for _, tt := range tests {
		prog := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(prog)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.ByteCode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedElement()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case *object.Nil:
		if actual != NIL {
			t.Errorf("object is not Nil. %T (%+v)", actual, actual)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not Array: %T (%+v)", actual, actual)
			return
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%q, got=%d", len(expected), len(array.Elements))
			return
		}

		for i, expElem := range expected {
			err := testIntegerObject(int64(expElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[object.HashKey]int64:
		hash, ok := actual.(*object.HashMap)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d",
				len(expected), len(hash.Pairs))
			return
		}

		for expKey, expVal := range expected {
			pair, ok := hash.Pairs[expKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}

			err := testIntegerObject(expVal, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case *object.Error:
		err, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error. got=%T (%+v)", actual, actual)
		}

		if err.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, err.Message)
		}
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTest{
		{"true", true},
		{"false", false},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTest{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 }", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", NIL},
		{"if (false) { 10 }", NIL},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTest{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two;", 3},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTest{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashMapLiterals(t *testing.T) {
	tests := []vmTest{
		{"{}", map[object.HashKey]int64{}},
		{
			"{1: 2, 3: 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 3}).HashKey(): 4,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTest{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", NIL},
		{"[1, 2, 3][99]", NIL},
		{"[1][-1]", NIL},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", NIL},
		{"{}[0]", NIL},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTest{
		{
			input:    "let fivePlusTen = fn() { 5 + 10 }; fivePlusTen();",
			expected: 15,
		},
		{
			input: `
			let one = fn() { 1; };
			let two = fn() { 2; };
			one() + two();`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { 1; };
			let b = fn() { a() + 1; };
			let c = fn() { b() + 1; };
			c()`,
			expected: 3,
		},
		{
			input: `
			let early = fn() { return 99; 100; };
			early();`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTest{
		{
			input: `
			let noRet = fn() {}; noRet();`,
			expected: NIL,
		},
		{
			input: `
			let noRet1 = fn() { };
			let noRet2 = fn() { noRet1(); };
			noRet1();
			noRet2();
			`,
			expected: NIL,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTest{
		{
			input:    `let one = fn() { let one = 1; one }; one();`,
			expected: 1,
		},
		{
			input: `let oneAndTwo = fn() { let one = 1; let two = 2; one + two };
			oneAndTwo();`,
			expected: 3,
		},
		{
			input: `let oneAndTwo = fn() { let one = 1; let two = 2; one + two };
			let threeAndFour = fn() { let three = 3; let four = 4; three + four };
			oneAndTwo() + threeAndFour();`,
			expected: 10,
		},
		{
			input: `let firstFoo = fn() { let foo = 50; foo }
			let secondFoo = fn() { let foo = 100; foo }
			firstFoo() + secondFoo()`,
			expected: 150,
		},
		{
			input: `let globalSeed = 50;
			let minusOne = fn() {
				let num = 1;
				globalSeed - num
			}
			let minusTwo = fn() {
				let num = 2;
				globalSeed - num
			}
			minusOne() + minusTwo()
			`,
			expected: 97,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTest{
		{
			input:    `let id = fn(a) { a }; id(4)`,
			expected: 4,
		},
		{
			input:    `let sum = fn(a, b) { a + b }; sum(1, 2);`,
			expected: 3,
		},
		{
			input:    `let sum = fn(a, b) { let c = a + b; c }; sum(1, 2);`,
			expected: 3,
		},
		{
			input:    `let sum = fn(a, b) { let c = a + b; c }; sum(1, 2) + sum(3, 4);`,
			expected: 10,
		},
		{
			input: `
			let sum = fn(a, b) { let c = a + b; c };
			let outer = fn() { sum(1, 2) + sum(3, 4) }; outer()`,
			expected: 10,
		},
		{
			input: `
			let globalNum = 10;
			let sum = fn(a, b) {
				let c = a + b;
				c + globalNum
			}
			let outer = fn() {
				sum(1, 2) + sum(3, 4) + globalNum;
			}
			
			outer() + globalNum;`,
			expected: 50,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTest{
		{
			input:    "fn() { 1; }(1)",
			expected: "wrong number of arguments: want=0, got=1",
		},
		{
			input:    "fn(a) { a; }()",
			expected: "wrong number of arguments: want=1, got=0",
		},
		{
			input:    "fn(a, b) { a + b; }(1)",
			expected: "wrong number of arguments: want=2, got=1",
		},
	}

	for _, tt := range tests {
		prog := parse(tt.input)
		comp := compiler.New()

		err := comp.Compile(prog)
		if err != nil {
			t.Errorf("compiler error: %s", err)
		}

		vm := New(comp.ByteCode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none.")
		}

		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q\ngot=%q", tt.expected, err)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTest{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{
				Message: "argument to `len` is not supported, got INTEGER",
			},
		},
		{
			`last([1, 2, 3]);`,
			3,
		},
		{
			`len("one", "two")`,
			&object.Error{
				Message: "wrong number of arguments. got=2, want=1",
			},
		},
		{`len([])`, 0},
		{`len([1, 2, 3, 4])`, 4},
		{`first([1, 2, 3])`, 1},
		{`first([])`, NIL},
		{`puts("hello, world")`, NIL},
		{
			`first(1)`,
			&object.Error{
				Message: "argument to `first` must be ARRAY, got INTEGER",
			},
		},
		{`last([])`, NIL},
		{
			`last(1)`,
			&object.Error{
				Message: "argument to `last` must be ARRAY, got INTEGER",
			},
		},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, NIL},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`,
			&object.Error{
				Message: "argument to `push` must be ARRAY, got INTEGER",
			},
		},
	}

	runVmTests(t, tests)
}
