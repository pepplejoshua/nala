package vm

import (
	"nala/object"
	"nala/opcode"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *object.Closure, bp int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1, // stores instruction pointer for current frame
		basePointer: bp, // stores the stack pointer's former position for return
		// after a function
	}
}

func (f *Frame) Instructions() opcode.Instructions {
	return f.cl.Fn.Instructions
}
