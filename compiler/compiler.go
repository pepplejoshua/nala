package compiler

import (
	"fmt"
	"nala/ast"
	"nala/object"
	"nala/opcode"
)

type EmittedInstruction struct {
	OpCode   opcode.OpCode
	Position int
}

type Compiler struct {
	instructions opcode.Instructions
	constants    []object.Object // handles language constants (integers and other objects)
	symbolTable  *SymbolTable    // handles identifier bindings

	recentInstruction   EmittedInstruction
	previousInstruction EmittedInstruction
}

type ByteCode struct {
	Instructions opcode.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions:        opcode.Instructions{},
		constants:           []object.Object{},
		symbolTable:         NewSymbolTable(),
		recentInstruction:   EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
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
		case ">": // can elide OpLThan into OpGThan by compiling in reverse (right before left)
			// and then pushing OPGThan onto the stack
			c.emit(opcode.OpGThan)
		case "<":
			c.emit(opcode.OpLThan)
		case "==":
			c.emit(opcode.OpEqual)
		case "!=": // can elide into combo of ! and ==. First compile like normal, then push
			// OpEqual and then OpNegateBool
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
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(opcode.OpSetGlobal, symbol.Index)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(opcode.OpGetGlobal, symbol.Index)
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		// we will similarly use the EmittedInstructions object
		// to backtrack and correct the address for OpJumpNotTruthy
		jmpNTPos := c.emit(opcode.OpJumpNotTruthy, 0)
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		// remove errant Pop statement (if it exists), so we can arbitrarily return
		// values like if (true) { 5 } should return a 5.
		if c.recentInstructionIsPop() {
			c.removeLastPop()
		}
		// insert jump to finish consequence section
		// return program to normal flow
		jmpPos := c.emit(opcode.OpJump, 9999)

		// fix the jmpNotTruthy address
		afterConsequencePos := len(c.instructions)
		// update jump location with correct address
		c.changeOperand(jmpNTPos, afterConsequencePos)

		if node.Alternative == nil {
			// correct jump location
			// all if instructions are Compiled
			// so we can set out jump location to after
			c.emit(opcode.OpNil)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.recentInstructionIsPop() {
				c.removeLastPop()
			}
		}

		afterAlternative := len(c.instructions)
		c.changeOperand(jmpPos, afterAlternative)
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(opcode.OpArray, len(node.Elements))
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

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := opcode.OpCode(c.instructions[opPos])
	newInstruction := opcode.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) recentInstructionIsPop() bool {
	return c.recentInstruction.OpCode == opcode.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.recentInstruction.Position]
	c.recentInstruction = c.previousInstruction
}

func (c *Compiler) emit(op opcode.OpCode, operands ...int) int {
	ins := opcode.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setRecentInstruction(op, pos)
	return pos
}

func (c *Compiler) setRecentInstruction(op opcode.OpCode, pos int) {
	prev := c.recentInstruction
	recent := EmittedInstruction{OpCode: op, Position: pos}

	c.previousInstruction = prev
	c.recentInstruction = recent
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
