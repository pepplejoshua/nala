package compiler

import (
	"fmt"
	"nala/ast"
	"nala/object"
	"nala/opcode"
)

type Compiler struct {
	instructions opcode.Instructions
	constants    []object.Object
}

type ByteCode struct {
	Instructions opcode.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: opcode.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(opcode.OpPop)

	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(opcode.OpAdd)
		case "-":
			c.emit(opcode.OpSubtract)
		case "/":
			c.emit(opcode.OpDivide)
		case "*":
			c.emit(opcode.OpMultiply)
		case "%":
			c.emit(opcode.OpModulo)
		case ">":
			c.emit(opcode.OpGThan)
		case "<":
			c.emit(opcode.OpLThan)
		case "==":
			c.emit(opcode.OpEqual)
		case "!=":
			c.emit(opcode.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "-":
			c.emit(opcode.OpNegateInt)
		case "!":
			c.emit(opcode.OpNegateBool)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(integer))
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(str))
	case *ast.Boolean:
		if node.Value {
			c.emit(opcode.OpTrue)
		} else {
			c.emit(opcode.OpFalse)
		}
	}
	return nil
}

func (c *Compiler) emit(op opcode.OpCode, operands ...int) int {
	ins := opcode.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(c.instructions) // returns the index of the just inserted instruction
	c.instructions = append(c.instructions, ins...)
	return posNewIns
}

func (c *Compiler) addConstant(obj object.Object) int {
	ref, exists := c.isExistingConstant(obj)
	if !exists {
		c.constants = append(c.constants, obj)
		return len(c.constants) - 1
	} else {
		return ref
	}
}

func (c *Compiler) isExistingConstant(obj object.Object) (int, bool) {
	hashAble, ok := obj.(object.Hashable)
	for i, con := range c.constants {
		cHash, cOk := con.(object.Hashable)
		if ok && cOk {
			if hashAble.HashKey() == cHash.HashKey() {
				return i, true
			}
		} else {
			if obj.Inspect() == con.Inspect() {
				return i, true
			}
		}
	}
	return -1, false
}
