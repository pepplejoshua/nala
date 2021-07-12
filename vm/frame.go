package vm

import (
	"nala/object"
	"nala/opcode"
)

type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, bp int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1, // stores instruction pointer for current frame
		basePointer: bp, // stores the stack pointer's former position for return
		// after a function
	}
}

func (f *Frame) Instructions() opcode.Instructions {
	return f.fn.Instructions
}
