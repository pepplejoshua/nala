package vm

import (
	"nala/object"
	"nala/opcode"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{
		fn: fn,
		ip: -1,
	}
}

func (f *Frame) Instructions() opcode.Instructions {
	return f.fn.Instructions
}
