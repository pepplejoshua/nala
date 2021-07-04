package evaluator

import (
	"fmt"
	"nala/ast"
	"nala/object"
)

var (
	NIL   = &object.Nil{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {

	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatements(node.Statements, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isErrorObj(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isErrorObj(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return getBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isErrorObj(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isErrorObj(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isErrorObj(right) {
			return right
		}
		return evalInfixExpression(left, node.Operator, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	}

	return nil
}

// handles top level evaluation of program
func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var res object.Object

	for _, stmt := range stmts {
		res = Eval(stmt, env)

		// if we have found a return or error statement,
		// we want to return prematurely without eval of remaining statements
		switch res := res.(type) {
		case *object.ReturnValue:
			return res.Value
		case *object.Error:
			return res
		}
	}
	return res
}

// handles evaluating BlockStatements (nested or not) with potential return statements
// if it runs into a return statement, it returns it.
// top level evalProgram will unwrap it and return to user.
func evalBlockStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var res object.Object

	for _, stmt := range stmts {
		res = Eval(stmt, env)

		// if we have found a return statement, we want to return prematurely without
		// eval of remaining statements
		if res != nil {
			resType := res.Type()
			if resType == object.RETURN_VALUE_OBJ || resType == object.ERROR_OBJ {
				return res
			}
		}
	}
	return res
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	cond := Eval(node.Condition, env)
	if isErrorObj(cond) {
		return cond
	}

	if isTruthy(cond) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	} else {
		return NIL
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}

	return val
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangPrefixExpression(right)
	case "-":
		return evalMinusPrefixExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}

}

func evalBangPrefixExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NIL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixExpression(right object.Object) object.Object {
	if !typeChecks(right.Type(), object.INTEGER_OBJ) {
		return newError("unknown operator: -%s", right.Type())
	}

	val := right.(*object.Integer).Value
	return &object.Integer{Value: -val}
}

func evalInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch {
	case operandTypeChecks(left.Type(), right.Type(), object.INTEGER_OBJ):
		return evalIntegerInfixExpression(left, operator, right)
	case operandTypeChecks(left.Type(), right.Type(), object.BOOLEAN_OBJ):
		return evalBooleanInfixExpression(left, operator, right)
	default:
		if left.Type() != right.Type() {
			return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch operator {
	case "+":
		return evalIntAdditionInfixExpression(left, right)
	case "-":
		return evalIntMinusInfixExpression(left, right)
	case "*":
		return evalIntMultiplicationInfixExpression(left, right)
	case "/":
		return evalIntDivisionInfixExpression(left, right)
	case "%":
		return evalIntModuloInfixExpression(left, right)
	case ">":
		return evalIntGTInfixExpression(left, right)
	case "<":
		return evalIntLTInfixExpression(left, right)
	case "==":
		return evalIntEqualityInfixExpression(left, right)
	case "!=":
		return evalIntInequalityInfixExpression(left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntAdditionInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return &object.Integer{Value: lval + rval}
}

func evalIntMinusInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return &object.Integer{Value: lval - rval}
}

func evalIntMultiplicationInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return &object.Integer{Value: lval * rval}
}

func evalIntDivisionInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value

	if rval == 0 {
		return newError("division by Zero: %d / 0", lval)
	}
	return &object.Integer{Value: lval / rval}
}

func evalIntModuloInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value

	if rval == 0 {
		return newError("modulo by Zero: %d %% 0", lval)
	}
	return &object.Integer{Value: lval % rval}
}

func evalIntGTInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return getBooleanObject(lval > rval)
}

func evalIntLTInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return getBooleanObject(lval < rval)
}

func evalIntEqualityInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return getBooleanObject(lval == rval)
}

func evalIntInequalityInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return getBooleanObject(lval != rval)
}

func evalBooleanInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch operator {
	case "==":
		return evalBoolEqualityInfixExpression(left, right)
	case "!=":
		return evalBoolInequalityInfixExpression(left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBoolEqualityInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Boolean).Value
	rval := right.(*object.Boolean).Value
	return getBooleanObject(lval == rval)
}

func evalBoolInequalityInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.Boolean).Value
	rval := right.(*object.Boolean).Value
	return getBooleanObject(lval != rval)
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
	switch value {
	case NIL:
		return false
	case FALSE:
		return false
	default:
		return true
	}
}

func getBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func typeChecks(typeC object.ObjectType, expected object.ObjectType) bool {
	return typeC == expected
}

func operandTypeChecks(op1 object.ObjectType, op2 object.ObjectType, etype object.ObjectType) bool {
	return typeChecks(op1, etype) && typeChecks(op2, etype)
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isErrorObj(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
