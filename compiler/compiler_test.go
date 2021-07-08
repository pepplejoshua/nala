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
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "3 * 1",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 / 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpDivide),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "2 % 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
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
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func parse(in string) ast.Node {
	return parser.New(lexer.New(in)).ParseProgram()
}

func testConstants(expCons []interface{}, bcCons []object.Object) error {
	if len(expCons) != len(bcCons) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d", len(bcCons), len(expCons))
	}

	for i, cons := range expCons {
		switch cons := cons.(type) {
		case int:
			err := testIntegerObject(int64(cons), bcCons[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}

		}
	}
	return nil
}

func testInstructions(expIns []opcode.Instructions, bcIns opcode.Instructions) error {
	concat := concatInstructions(expIns)

	if len(bcIns) != len(concat) {
		return fmt.Errorf("wrong instruction length.\nwant=%q, got=%q", concat, bcIns)
	}

	for i, ins := range concat {
		if bcIns[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q", i, ins, bcIns[i])
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
	}

	runCompilerTests(t, tests)
}
