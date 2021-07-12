package compiler

import (
	"fmt"
	"nala/ast"
	"nala/object"
	"nala/opcode"
	"sort"
)

type EmittedInstruction struct {
	OpCode   opcode.OpCode
	Position int
}

type CompilationScope struct {
	instructions        opcode.Instructions // instruction to be returned in *object.CompiledFunction
	recentInstruction   EmittedInstruction  // recent instruction for this compilation scope
	previousInstruction EmittedInstruction  // instruction before recent for this compilation scope
}

type Compiler struct {
	constants   []object.Object // handles language constants (integers and other objects)
	symbolTable *SymbolTable    // handles identifier bindings

	scopes     []CompilationScope // slice allowing separate compilation of individual scoped objects (e.g Functions)
	scopeIndex int                // index of current scope of compilation
}

type ByteCode struct {
	Instructions opcode.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        opcode.Instructions{},
		recentInstruction:   EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
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
		Instructions: c.currentInstructions(),
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
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// checks for an implicit return value
		if c.recentInstructionIs(opcode.OpPop) {
			c.replaceRecentPopWithReturn() // this pop means there is an implicit return value
			// from an expression. let statements will have no pop
		}
		// the lack of the pop or ReturnValue shows theres no explicit or implicit return
		// so we put a Return OpCode in that code
		if !c.recentInstructionIs(opcode.OpReturnValue) {
			c.emit(opcode.OpReturn)
		}

		numLocals := c.symbolTable.numDefinitions // number of locals defined in this scope
		instructions := c.leaveScope()
		compiledFn := &object.CompiledFunction{
			Instructions:    instructions,
			NumOfLocals:     numLocals,
			NumOfParameters: len(node.Parameters),
		}
		c.emit(opcode.OpConstant, c.addConstant(compiledFn))
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(opcode.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}
		c.emit(opcode.OpCall, len(node.Arguments))
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
		// by defining a location where the function can be found,
		// we allow recursion
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpSetGlobal, symbol.Index)
		} else if symbol.Scope == LocalScope {
			c.emit(opcode.OpSetLocal, symbol.Index)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpGetGlobal, symbol.Index)
		} else if symbol.Scope == LocalScope {
			c.emit(opcode.OpGetLocal, symbol.Index)
		} else if symbol.Scope == BuiltInScope {
			c.emit(opcode.OpGetBuiltin, symbol.Index)
		}
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
		if c.recentInstructionIs(opcode.OpPop) {
			c.removeRecentPop()
		}
		// insert jump to finish consequence section
		// return program to normal flow
		jmpPos := c.emit(opcode.OpJump, 9999)

		// fix the jmpNotTruthy address
		afterConsequencePos := len(c.currentInstructions())
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

			if c.recentInstructionIs(opcode.OpPop) {
				c.removeRecentPop()
			}
		}

		afterAlternative := len(c.currentInstructions())
		c.changeOperand(jmpPos, afterAlternative)
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(opcode.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		// collect and sort keys in ascending order
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(opcode.OpHashMap, len(node.Pairs)*2)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(opcode.OpIndex)
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
	op := opcode.OpCode(c.currentInstructions()[opPos])
	newInstruction := opcode.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	curIns := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		curIns[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) replaceRecentPopWithReturn() {
	recentPos := c.currentScope().recentInstruction.Position
	c.replaceInstruction(recentPos, opcode.Make(opcode.OpReturnValue))
	c.scopes[c.scopeIndex].recentInstruction.OpCode = opcode.OpReturnValue
}

func (c *Compiler) recentInstructionIs(op opcode.OpCode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.currentScope().recentInstruction.OpCode == op
}

func (c *Compiler) removeRecentPop() {
	recent := c.currentScope().recentInstruction
	prev := c.currentScope().previousInstruction

	old := c.currentInstructions()
	new := old[:recent.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].previousInstruction = prev
}

func (c *Compiler) emit(op opcode.OpCode, operands ...int) int {
	ins := opcode.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setRecentInstruction(op, pos)
	return pos
}

func (c *Compiler) setRecentInstruction(op opcode.OpCode, pos int) {
	prev := c.currentScope().recentInstruction
	recent := EmittedInstruction{OpCode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = prev
	c.scopes[c.scopeIndex].recentInstruction = recent
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(c.currentInstructions()) // returns the index of the just inserted instruction
	newIns := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = newIns
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

func (c *Compiler) currentInstructions() opcode.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) currentScope() CompilationScope {
	return c.scopes[c.scopeIndex]
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        opcode.Instructions{},
		recentInstruction:   EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() opcode.Instructions {
	ins := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer
	return ins
}

func (c *Compiler) Decompile(ins opcode.Instructions, constants []object.Object, globals []object.Object, offset string, depth int) {
	i := 0
	for i < len(ins) {
		def, err := opcode.Lookup(ins[i]) // get operand definition
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}
		operands, read := opcode.ReadOperands(def, ins[i+1:])

		fmt.Printf(offset+"%04d....%s....[%d bytes]\n", i, c.fmtInstruction(def, operands), i+read+1)
		c.showOperand(def, operands, constants, globals, offset+"            ", depth+1)
		i += read + 1
	}
}

func (c *Compiler) fmtInstruction(def *opcode.Definition, operands []int) string {
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

func (c *Compiler) showOperand(def *opcode.Definition, operands []int, constants, globals []object.Object, offset string, depth int) {
	opCount := len(def.OperandWidths)

	if opCount != len(operands) {
		fmt.Printf("ERROR: operand len %d does not match defined %d\n", len(operands), opCount)
	}
	if opCount == 0 {
		return
	}
	if len(constants) > 0 && len(globals) > 0 {
		if def.Name == "OpConstant" {
			Ind := operands[0]
			op := constants[Ind]
			switch op := op.(type) {
			case *object.Integer, *object.Boolean,
				*object.String, *object.Array, *object.HashMap, *object.Function:
				fmt.Println(offset + "[Constant: " + op.Inspect() + "]")
			case *object.CompiledFunction:
				c.Decompile(op.Instructions, constants, globals, offset, depth+1)
			}
		} else if def.Name == "OpGetGlobal" {
			op := globals[operands[0]]
			switch op := op.(type) {
			case *object.Integer, *object.Boolean,
				*object.String, *object.Array, *object.HashMap, *object.Function:
				fmt.Println(offset + "[Global: " + op.Inspect() + "]")
			case *object.CompiledFunction:
				if depth < 5 {
					c.Decompile(op.Instructions, constants, globals, offset, depth+1)
				}
			}
		} else if def.Name == "OpGetBuiltin" {
			builtin := object.Builtins[operands[0]]
			fmt.Println(offset + "[Builtin: " + builtin.Name + "]")
		}

	}

	// } else if def.Name == "OpGetGlobal" || def.Name == "OpSetGlobal"
}
