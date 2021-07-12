package compiler

import (
	"fmt"
	"nala/ast"
	"nala/lexer"
	"nala/object"
	"nala/opcode"
	"nala/parser"
	"testing"
)

type CompilerTest struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []opcode.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 - 2",
			expectedConstants: []interface{}{2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "3 * 1",
			expectedConstants: []interface{}{3, 1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 / 2",
			expectedConstants: []interface{}{2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpDivide),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 % 2",
			expectedConstants: []interface{}{2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpModulo),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1; 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 > 2",
			expectedConstants: []interface{}{2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpGThan),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 < 2",
			expectedConstants: []interface{}{2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpLThan),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "-20",
			expectedConstants: []interface{}{20},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpNegateInt),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []CompilerTest) {
	t.Helper()

	for _, tt := range tests {
		prog := parse(tt.input)

		compiler := New()
		err := compiler.Compile(prog)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.ByteCode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			fmt.Println(tt.input)
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(tt.expectedConstants, bytecode.Constants)
		if err != nil {
			fmt.Println(tt.input)
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func parse(in string) ast.Node {
	return parser.New(lexer.New(in)).ParseProgram()
}

func testConstants(expCons []interface{}, actual []object.Object) error {
	if len(expCons) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d", len(actual), len(expCons))
	}

	for i, cons := range expCons {
		switch cons := cons.(type) {
		case int:
			err := testIntegerObject(int64(cons), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case []opcode.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", i, actual[i])
			}

			err := testInstructions(cons, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
		}
	}
	return nil
}

func testInstructions(expIns []opcode.Instructions, bcIns opcode.Instructions) error {
	concat := concatInstructions(expIns)

	if len(bcIns) != len(concat) {
		return fmt.Errorf("wrong instruction length.\nwant=%q\ngot=%q", concat, bcIns)
	}

	for i, ins := range concat {
		if bcIns[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q", i, concat, bcIns)
		}
	}

	return nil
}

// concat slice of slice of bytes into a single slice of bytes
func concatInstructions(ins []opcode.Instructions) opcode.Instructions {
	out := opcode.Instructions{}

	for _, in := range ins {
		out = append(out, in...)
	}

	return out
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

func TestBooleanExpressions(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpGThan),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpLThan),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpNotEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "true == true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpNotEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "!false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpNegateBool),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpNegateBool),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

// each conditional ends with a Pop since it is an ExpressionStatement
func TestConditionals(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             "if (true) { 10; }; 3333;",
			expectedConstants: []interface{}{10, 3333},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),              // 0000 + 1
				opcode.Make(opcode.OpJumpNotTruthy, 10), // 0001 + 3
				opcode.Make(opcode.OpConstant, 0),       // 0004 + 3
				opcode.Make(opcode.OpJump, 11),          // 0007 + 3
				opcode.Make(opcode.OpNil),               // 0010 + 1
				opcode.Make(opcode.OpPop),               // 0011 + 1
				opcode.Make(opcode.OpConstant, 1),       // 0012 + 3
				opcode.Make(opcode.OpPop),               //0015 + 1
				// 0016
			},
		},
		{
			input:             "if (true) { 10; } else { 25 }; 3333;",
			expectedConstants: []interface{}{10, 25, 3333},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),              // 0000 + 1
				opcode.Make(opcode.OpJumpNotTruthy, 10), // 0001 + 3
				opcode.Make(opcode.OpConstant, 0),       // 0004 + 3
				opcode.Make(opcode.OpJump, 13),          // 0007 + 3
				opcode.Make(opcode.OpConstant, 1),       // 0010 + 3
				opcode.Make(opcode.OpPop),               // 0013 + 1
				opcode.Make(opcode.OpConstant, 2),       // 0014 + 3
				opcode.Make(opcode.OpPop),               //0017 + 1
				// 0018
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             `let one = 1; let two = 2;`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSetGlobal, 1),
			},
		},
		{
			input:             `let one = 1; one;`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             `let one = 1; let two = one; two`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpSetGlobal, 1),
				opcode.Make(opcode.OpGetGlobal, 1),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpArray, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpArray, 3),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpArray, 3),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashMapLiteral(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpHashMap, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpHashMap, 6),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpHashMap, 4),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []CompilerTest{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpArray, 3),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpIndex),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpHashMap, 2),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpIndex),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []CompilerTest{
		{
			input: "fn() { return 5 + 10 }",
			expectedConstants: []interface{}{
				5,
				10,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "fn() { 5 + 10 }",
			expectedConstants: []interface{}{
				5,
				10,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "fn() { 1; 2 } ",
			expectedConstants: []interface{}{
				1,
				2,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "fn() { }",
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpReturn),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	comp := New()
	if comp.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 0)
	}
	globalSymTable := comp.symbolTable
	comp.emit(opcode.OpMultiply)

	comp.enterScope()
	if comp.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 1)
	}
	comp.emit(opcode.OpSubtract)
	if len(comp.scopes[comp.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopeIndex].instructions))
	}
	recent := comp.scopes[comp.scopeIndex].recentInstruction
	if recent.OpCode != opcode.OpSubtract {
		t.Errorf("recentInstruction.OpCode wrong. got=%d, want=%d", recent.OpCode, opcode.OpSubtract)
	}

	if comp.symbolTable.Outer != globalSymTable {
		t.Errorf("compiler did not enclose SymbolTable")
	}
	comp.leaveScope()

	if comp.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 0)
	}
	if comp.symbolTable != globalSymTable {
		t.Errorf("compiler did not restore SymbolTable")
	}
	if comp.symbolTable.Outer != nil {
		t.Errorf("compiler modified SymbolTable incorrectly")
	}
	comp.emit(opcode.OpAdd)
	if len(comp.scopes[comp.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopeIndex].instructions))
	}
	recent = comp.scopes[comp.scopeIndex].recentInstruction
	if recent.OpCode != opcode.OpAdd {
		t.Errorf("recentInstruction.OpCode wrong. got=%d, want=%d", recent.OpCode, opcode.OpSubtract)
	}
	previous := comp.scopes[comp.scopeIndex].previousInstruction
	if previous.OpCode != opcode.OpMultiply {
		t.Errorf("previousInstruction.OpCode wrong. got=%d, want=%d", previous.OpCode, opcode.OpMultiply)
	}
}

func TestFunctionCalls(t *testing.T) {
	tests := []CompilerTest{
		{
			input: "fn() {24}()",
			expectedConstants: []interface{}{
				24,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpCall, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "let noArgs = fn() { 24 }; noArgs()",
			expectedConstants: []interface{}{
				24,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpCall, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "let oneArgs = fn(a) { a }; oneArgs(24)",
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
				24,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpCall, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `let manyArgs = fn(a, b, c, d) { a; b; c; d; }; manyArgs(1, 2, 3, 4)`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 1),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 2),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 3),
					opcode.Make(opcode.OpReturnValue),
				},
				1,
				2,
				3,
				4,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpCall, 4),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalLetStatementScopes(t *testing.T) {
	tests := []CompilerTest{
		{
			input: `let n = 55; fn() { n }`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetGlobal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `fn() { let num = 55; num }`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `fn() { let a = 55; let b = 77; a + b }`,
			expectedConstants: []interface{}{
				55,
				77,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpSetLocal, 1),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpGetLocal, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}
