package vm

import (
	"fmt"
	"nala/compiler"
	"nala/object"
	"nala/opcode"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var (
	TRUE = &object.Boolean{
		Value:       true,
		HashableKey: &object.HashKey{},
	}
	FALSE = &object.Boolean{
		Value:       false,
		HashableKey: &object.HashKey{},
	}
	NIL = &object.Nil{}
)

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack []object.Object
	sp    int // ALways points to the next value Top of stack is stack[sp-1]

	frames      []*Frame // Call Stack to contain Frames of called functions
	framesIndex int      // Index into the call stack
}

func (vm *VM) Globals() []object.Object {
	return vm.globals
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedElement() object.Object {
	return vm.stack[vm.sp]
}

// rewrite switch into a dispatch map of functions
func (vm *VM) Run() error {
	var insPtr int
	var ins opcode.Instructions
	var op opcode.OpCode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		insPtr = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = opcode.OpCode(ins[insPtr]) // fetch instruction

		switch op { // decode instruction
		case opcode.OpConstant:
			constIndex := opcode.ReadUInt16(ins[insPtr+1:]) // for accessing constants pool

			// advance instruction pointer
			vm.currentFrame().ip += 2

			// execute instruction
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case opcode.OpAdd, opcode.OpSubtract,
			opcode.OpMultiply, opcode.OpDivide, opcode.OpModulo,
			opcode.OpGThan, opcode.OpLThan, opcode.OpEqual, opcode.OpNotEqual:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case opcode.OpPop:
			vm.pop()
		case opcode.OpTrue:
			err := vm.push(TRUE)
			if err != nil {
				return err
			}
		case opcode.OpFalse:
			err := vm.push(FALSE)
			if err != nil {
				return err
			}
		case opcode.OpNegateBool, opcode.OpNegateInt:
			err := vm.executeUnaryOperation(op)
			if err != nil {
				return err
			}
		case opcode.OpJump:
			newPos := int(opcode.ReadUInt16(ins[insPtr+1:]))
			vm.currentFrame().ip = newPos - 1 // execute the jump
		case opcode.OpJumpNotTruthy:
			newPos := int(opcode.ReadUInt16(ins[insPtr+1:]))
			vm.currentFrame().ip += 2 // we have read 2 bytes (16 bits)

			cond := vm.pop()
			if !isTruthy(cond) {
				vm.currentFrame().ip = newPos - 1 // perform JumpNotTruthy, else continue executing
			}
		case opcode.OpNil:
			err := vm.push(NIL)
			if err != nil {
				return err
			}
		case opcode.OpSetGlobal:
			globalIndex := opcode.ReadUInt16(ins[insPtr+1:])

			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop() // set the global at that position to the top of stack
		case opcode.OpGetGlobal:
			globalIndex := opcode.ReadUInt16(ins[insPtr+1:])

			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case opcode.OpArray:
			numElems := int(opcode.ReadUInt16(ins[insPtr+1:]))
			vm.currentFrame().ip += 2

			start := vm.sp - numElems
			array := vm.buildArray(start, vm.sp)
			vm.sp = start
			err := vm.push(array)
			if err != nil {
				return err
			}
		case opcode.OpHashMap:
			numElems := int(opcode.ReadUInt16(ins[insPtr+1:]))
			vm.currentFrame().ip += 2

			start := vm.sp - numElems
			hashMap, err := vm.buildHashMap(start, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = start
			err = vm.push(hashMap)
			if err != nil {
				return err
			}
		case opcode.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case opcode.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(NIL)
			if err != nil {
				return err
			}
		case opcode.OpReturnValue:
			returnVal := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnVal)
			if err != nil {
				return err
			}
		case opcode.OpCall:
			// get number of arguments for the function call
			numArgs := int(opcode.ReadUInt8(ins[insPtr+1:]))
			vm.currentFrame().ip++

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}
		case opcode.OpSetLocal:
			localIndex := opcode.ReadUInt8(ins[insPtr+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			// use the frame's base pointer as an offset into the stack + the local's index
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case opcode.OpGetLocal:
			localIndex := opcode.ReadUInt8(ins[insPtr+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			// use the frame's base pointer as an offset into the stack + the local's index
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case opcode.OpGetBuiltin:
			builtinIndex := opcode.ReadUInt8(ins[insPtr+1:])
			vm.currentFrame().ip++

			def := object.Builtins[builtinIndex]

			err := vm.push(def.BuiltIn)
			if err != nil {
				return err
			}
		case opcode.OpClosure:
			constIndex := opcode.ReadUInt16(ins[insPtr+1:])      // read location of compiled func
			numOfFreeSymbols := opcode.ReadUInt8(ins[insPtr+3:]) // read number of free variables
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), int(numOfFreeSymbols))
			if err != nil {
				return err
			}
		case opcode.OpGetFree:
			freeIndex := opcode.ReadUInt8(ins[insPtr+1:])
			vm.currentFrame().ip++
			currentCl := vm.currentFrame().cl
			println("free index: ", freeIndex)
			free := currentCl.FreeVariables[freeIndex]
			// fmt.Println(currentCl.FreeVariables)
			err := vm.push(free)
			if err != nil {
				return err
			}

		}
		def, _ := opcode.Lookup(byte(op))
		fmt.Println(def.Name)
		fmt.Println("*************")
		fmt.Println("st::", vm.stack[0:vm.sp])
		fmt.Println("cs::", vm.constants)
		// fmt.Println("gs::", vm.globals)
		fmt.Print("*************\n\n")
	}

	return nil
}

func (vm *VM) pushClosure(fnIndex int, freeSyms int) error {
	fnConst := vm.constants[fnIndex]
	fn, ok := fnConst.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", fnConst)
	}

	free := make([]object.Object, freeSyms)
	// read off free symbols from stack
	// fmt.Println(vm.stack[vm.sp-freeSyms : vm.sp])

	for i := 0; i < freeSyms; i++ {
		free[i] = vm.stack[vm.sp-freeSyms+i]
	}
	vm.sp = vm.sp - freeSyms

	closure := &object.Closure{
		Fn:            fn,
		FreeVariables: free,
	}
	return vm.push(closure)
}

func (vm *VM) executeCall(numArgs int) error {
	// reach down and get the function past the arguments
	callable := vm.stack[vm.sp-numArgs-1]
	println("function found at: ", vm.sp-numArgs-1)
	// fmt.Println(callable.Inspect(), " ", callable.Type())

	// fmt.Println(vm.currentFrame().Instructions().String())
	// fmt.Println("here: ", vm.currentFrame().ip-1)
	switch callable := callable.(type) {
	case *object.Closure:
		return vm.callClosure(callable, numArgs)
	case *object.BuiltIn:
		return vm.callBuiltin(callable, numArgs)
	default:
		return fmt.Errorf("calling non-closure and non-builtin [%T]", callable)
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumOfParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.Fn.NumOfParameters, numArgs)
	}
	frame := NewFrame(cl, vm.sp-numArgs) // move the basePointer even lower to include Arguments
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumOfLocals // this creates the hole
	// to store and get local variables on the stack
	return nil
}

func (vm *VM) callBuiltin(bi *object.BuiltIn, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	res := bi.Fn(args...)
	vm.sp = vm.sp - numArgs - 1 // move sp back to return position after function

	err := vm.push(res)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) buildArray(start, end int) object.Object {
	elems := make([]object.Object, end-start)

	for i := start; i < end; i++ {
		elems[i-start] = vm.stack[i]
	}

	return &object.Array{Elements: elems}
}

func (vm *VM) buildHashMap(start, end int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		val := vm.stack[i+1]

		pair := object.HashPair{
			Key:   key,
			Value: val,
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as a hash key: %s", key.Type())
		}
		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.HashMap{
		Pairs: hashedPairs,
	}, nil
}

func (vm *VM) executeUnaryOperation(op opcode.OpCode) error {
	right := vm.pop()

	switch r := right.(type) {
	case *object.Integer:
		if op != opcode.OpNegateInt {
			return fmt.Errorf("unknown integer operator: %d", op)
		} else {
			res := -r.Value
			return vm.push(&object.Integer{Value: res})
		}
	case *object.Boolean:
		if op != opcode.OpNegateBool {
			return fmt.Errorf("unknown boolean operator: %d", op)
		} else {
			res := !r.Value
			return vm.push(&object.Boolean{Value: res})
		}
	default:
		return fmt.Errorf("unsupported type %s for unary operation", r.Type())
	}
}

func (vm *VM) executeBinaryOperation(op opcode.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	switch lVal := left.(type) {
	case *object.Integer:
		rVal, ok := right.(*object.Integer)
		if !ok {
			return fmt.Errorf("disjointed types for operators: %s, %s", left.Type(), right.Type())
		}
		return vm.executeIntegerBinaryOperation(op, lVal.Value, rVal.Value)
	case *object.String:
		rVal, ok := right.(*object.String)
		if !ok {
			return fmt.Errorf("disjointed types for operators: %s, %s", left.Type(), right.Type())
		}

		return vm.executeStringBinaryOperation(op, lVal.Value, rVal.Value)
	case *object.Boolean:
		rVal, ok := right.(*object.Boolean)
		if !ok {
			return fmt.Errorf("disjointed types for operators: %s, %s", left.Type(), right.Type())
		}
		return vm.executeBooleanBinaryOperation(op, lVal.Value, rVal.Value)
	default:
		return fmt.Errorf("unsupported types %s and %s for binary operation", left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerBinaryOperation(op opcode.OpCode, left, right int64) error {
	var res interface{}
	switch op {
	case opcode.OpAdd:
		res = left + right
	case opcode.OpSubtract:
		res = left - right
	case opcode.OpModulo:
		if right != 0 {
			res = left % right
		} else {
			return fmt.Errorf("division by 0 error")
		}
	case opcode.OpDivide:
		if right != 0 {
			res = left / right
		} else {
			return fmt.Errorf("division by 0 error")
		}
	case opcode.OpMultiply:
		res = left * right
	case opcode.OpLThan:
		res = left < right
	case opcode.OpGThan:
		res = left > right
	case opcode.OpEqual:
		res = left == right
	case opcode.OpNotEqual:
		res = left != right
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	switch res := res.(type) {
	case int64:
		return vm.push(&object.Integer{Value: res})
	case bool:
		if res {
			return vm.push(TRUE)
		} else {
			return vm.push(FALSE)
		}
	}
	return nil
}

func (vm *VM) executeBooleanBinaryOperation(op opcode.OpCode, left, right bool) error {
	var res bool
	switch op {
	case opcode.OpEqual:
		res = left == right
	case opcode.OpNotEqual:
		res = left != right
	default:
		return fmt.Errorf("unknown boolean operator: %d", op)
	}
	return vm.push(&object.Boolean{Value: res})
}

func (vm *VM) executeStringBinaryOperation(op opcode.OpCode, left, right string) error {
	var res interface{}
	switch op {
	case opcode.OpAdd:
		res = left + right
	case opcode.OpEqual:
		res = left == right
	case opcode.OpNotEqual:
		res = left != right
	default:
		return fmt.Errorf("unknown boolean operator: %d", op)
	}
	switch res := res.(type) {
	case string:
		return vm.push(&object.String{Value: res})
	case bool:
		if res {
			return vm.push(TRUE)
		} else {
			return vm.push(FALSE)
		}
	}
	return nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASHMAP_OBJ:
		return vm.executeHashMapIndex(left, index)
	default:
		return fmt.Errorf("index operator not supportedL %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(left, index object.Object) error {
	arrObj := left.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrObj.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(NIL)
	}

	return vm.push(arrObj.Elements[i])
}

func (vm *VM) executeHashMapIndex(left, index object.Object) error {
	hashObj := left.(*object.HashMap)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(NIL)
	}

	return vm.push(pair.Value)
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func New(bc *compiler.ByteCode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bc.Instructions}
	mainClosure := &object.Closure{
		Fn: mainFn,
	}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bc.Constants,
		globals:     make([]object.Object, GlobalsSize),
		stack:       make([]object.Object, StackSize),
		sp:          0,
		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bc *compiler.ByteCode, globs []object.Object) *VM {
	vm := New(bc)
	vm.globals = globs
	return vm
}

func isTruthy(value object.Object) bool {
	switch value.Type() {
	case object.INTEGER_OBJ:
		switch value.Inspect() {
		case "0":
			return false
		default:
			return true
		}
	case object.BOOLEAN_OBJ:
		switch value {
		case TRUE:
			return true
		case FALSE:
			return false
		}
	case object.NIL_OBJ:
		return false
	default:
		return true
	}
	return false
}
