package opcode

import "testing"

type OpCodeTest struct {
	op       OpCode
	operands []int
	expected []byte
}

type ReadOperandsTest struct {
	op        OpCode
	operands  []int
	bytesRead int
}

func TestMake(t *testing.T) {
	tests := []OpCodeTest{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
		{OpClosure, []int{65534, 255}, []byte{byte(OpClosure), 255, 254, 255}},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)

		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d", len(tt.expected), len(instruction))
		}

		for i, b := range tt.expected {
			if instruction[i] != tt.expected[i] {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d", i, b, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 30000),
		Make(OpAdd),
		Make(OpGetLocal, 1),
		Make(OpClosure, 65535, 255),
	}

	exp := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 30000
0009 OpAdd
0010 OpGetLocal 1
0012 OpClosure 65535 255
`

	concat := Instructions{}

	for _, ins := range instructions {
		concat = append(concat, ins...)
	}

	// can rename string method to mini disassembler
	if concat.String() != exp {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q", exp, concat.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []ReadOperandsTest{
		{OpConstant, []int{65535}, 2},
		{OpGetLocal, []int{255}, 1},
		{OpClosure, []int{65535, 255}, 3},
	}

	for _, tt := range tests {
		ins := Make(tt.op, tt.operands...)

		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("opcode definition not found: %q\n", err)
		}

		opsRead, n := ReadOperands(def, ins[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. got=%d, want=%d", n, tt.bytesRead)
		}

		for i, exp := range tt.operands {
			if opsRead[i] != exp {
				t.Errorf("operand wrong. want=%d, got=%d", exp, opsRead[i])
			}
		}
	}
}
