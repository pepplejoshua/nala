package opcode

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Bytecode is a bunch of instructions encoded as bytes.
// Each Instruction contains an opcode along with optional operands.
// operands are encoded in BigEndian Order
type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i]) // Lookup operator to get operand widths
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d....%s....[%d bytes]\n", i, ins.fmtInstruction(def, operands), i+read+1)

		i += read + 1
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	opCount := len(def.OperandWidths)

	if opCount != len(operands) {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), opCount)
	}

	switch opCount {
	case 0:
		return def.Name
	case 1:
		if def.Name == "OpJump" || def.Name == "OpJumpNotTruthy" {
			return fmt.Sprintf("%s %04d", def.Name, operands[0])
		} else {
			return fmt.Sprintf("%s %d", def.Name, operands[0])
		}
	}

	return fmt.Sprintf("ERROR: unhandled opCount for %s\n", def.Name)
}

// Each opcode is a byte
type OpCode byte

type Definition struct {
	Name          string
	OperandWidths []int // how many bytes each operand takes up
}

// iota increments all the OpCodes by 1 each time
const (
	OpConstant OpCode = iota // used to push a constant from ConstantsPool[] onto stack.
	OpPop
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpModulo
	OpGThan
	OpLThan
	OpEqual
	OpNotEqual
	OpTrue
	OpFalse
	OpNegateInt
	OpNegateBool
	OpJumpNotTruthy
	OpJump
	OpNil
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHashMap
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
)

var definitions = map[OpCode]*Definition{
	OpConstant:      {"OpConstant", []int{2}}, // since the operand for OpConstant is 2, we can only reference 65536 constants (counting 0)
	OpAdd:           {"OpAdd", []int{}},
	OpSubtract:      {"OpSubtract", []int{}},
	OpMultiply:      {"OpMultiply", []int{}},
	OpDivide:        {"OpDivide", []int{}},
	OpModulo:        {"OpModulo", []int{}},
	OpGThan:         {"OpGreaterThan", []int{}},
	OpLThan:         {"OpLessThan", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpNegateBool:    {"OpNegateBool", []int{}},
	OpNegateInt:     {"OpNegateInt", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, // this sets a limit of Instruction 0-65534 jump addresses
	OpJump:          {"OpJump", []int{2}},          // so they are both 16 bits wide
	OpNil:           {"OpNil", []int{}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHashMap:       {"OpHashMap", []int{2}},
	OpIndex:         {"OpIndex", []int{}},
	OpCall:          {"OpCall", []int{}},
	OpReturnValue:   {"OpReturnValue", []int{}},
	OpReturn:        {"OpReturn", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[OpCode(op)] // look up OpCode definition from byte
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// turning OpCode and integer operands into an array of bytes
func Make(op OpCode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1 // 1 byte to encode the OpCode
	for _, w := range def.OperandWidths {
		instructionLen += w // then add on the additional widths for the operands
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1                  // offset is the number of bytes to start inserting after, which is 1 for right after our OpCode
	for i, o := range operands { // for each of the passed operands
		width := def.OperandWidths[i] // collect the appropriate width

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:offset+width], uint16(o))
		}
		offset += width // update offset to now be past the recently added operand
	}
	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUInt16(ins[offset : offset+width]))
		}
		offset += width // update offset to now be past the recently added operand
	}

	return operands, offset
}

func ReadUInt16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
