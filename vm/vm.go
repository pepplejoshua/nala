package vm

import (
	"fmt"
	"nala/compiler"
	"nala/object"
	"nala/opcode"
)

const StackSize = 2048

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
	constants    []object.Object
	instructions opcode.Instructions

	stack []object.Object
	sp    int // ALways points to the next value Top of stack is stack[sp-1]
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
	for insPtr := 0; insPtr < len(vm.instructions); insPtr++ {
		op := opcode.OpCode(vm.instructions[insPtr]) // fetch instruction

		switch op { // decode instruction
		case opcode.OpConstant:
			constIndex := opcode.ReadUInt16(vm.instructions[insPtr+1:]) // for accessing constants pool

			// advance instruction pointer
			insPtr += 2

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
			newPos := int(opcode.ReadUInt16(vm.instructions[insPtr+1:]))
			insPtr = newPos - 1 // execute the jump
		case opcode.OpJumpNotTruthy:
			newPos := int(opcode.ReadUInt16(vm.instructions[insPtr+1:]))
			insPtr += 2 // we have read 2 bytes (16 bits)

			cond := vm.pop()
			if !isTruthy(cond) {
				insPtr = newPos - 1 // perform JumpNotTruthy, else continue executing
			}
		case opcode.OpNil:
			err := vm.push(NIL)
			if err != nil {
				return err
			}
		}

	}

	return nil
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

func New(bc *compiler.ByteCode) *VM {
	return &VM{
		constants:    bc.Constants,
		instructions: bc.Instructions,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
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
