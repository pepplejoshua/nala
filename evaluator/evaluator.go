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
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		elems := evalExpressions(node.Elements, env)
		if len(elems) == 1 && isErrorObj(elems[0]) {
			return elems[0]
		}
		return &object.Array{Elements: elems}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isErrorObj(left) {
			return left
		}
		indx := Eval(node.Index, env)
		if isErrorObj(indx) {
			return indx
		}
		return evalIndexExpression(left, indx)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
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
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		if node.Function.TokenLiteral() == "quote" {
			// this freezes the object (does not interprete it)
			return quote(node.Arguments[0])
		}

		fn := Eval(node.Function, env)
		if isErrorObj(fn) {
			return fn
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isErrorObj(args[0]) {
			return args[0]
		}

		return applyFunction(fn, args)
	}

	return nil
}

func applyFunction(function object.Object, args []object.Object) object.Object {
	switch fn := function.(type) {
	case *object.Function:
		// this creates a static environment binding, as fn.env is the lexical env
		// from when it was defined vs whatever the current env is at the point of this call.
		// passing that env instead would be dynamic environment binding
		extendedEnv := extendFunctionEnv(fn, args)
		evald := Eval(fn.Body, extendedEnv)

		// do parameter counting to make sure right number of arguments were passed
		if len(args) != len(fn.Parameters) {
			err := "wrong number of arguments. got=%d, want=%d"
			return newError(err, len(args), len(fn.Parameters))
		}
		if retVal, ok := evald.(*object.ReturnValue); ok {
			return retVal.Value
		}
		return evald
	case *object.BuiltIn:
		// call the builtin
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramI, param := range fn.Parameters {
		env.Set(param.Value, args[paramI])
	}

	return env
}

func evalIndexExpression(left object.Object, index object.Object) object.Object {
	switch {
	case typeChecks(left.Type(), object.ARRAY_OBJ) &&
		typeChecks(index.Type(), object.INTEGER_OBJ):
		return evalArrayIndexExpression(left, index)
	case typeChecks(left.Type(), object.HASHMAP_OBJ):
		return evalHashMapIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array object.Object, index object.Object) object.Object {
	arr := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arr.Elements) - 1)

	if idx < 0 || idx > max {
		return NIL
	}
	return arr.Elements[idx]
}

func evalHashMapIndexExpression(hashObj object.Object, index object.Object) object.Object {
	hmap := hashObj.(*object.HashMap)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hmap.Pairs[key.HashKey()]
	if !ok {
		return NIL
	}

	return pair.Value
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isErrorObj(key) {
			return key
		}

		hashedKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valNode, env)
		if isErrorObj(value) {
			return value
		}
		hsh := hashedKey.HashKey()
		pairs[hsh] = object.HashPair{Key: key, Value: value}
	}

	return &object.HashMap{Pairs: pairs}
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

// this handles evaluating a function call's arguments (Expressions)
// and returns error if any arises
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var res []object.Object

	for _, e := range exps {
		evald := Eval(e, env)
		if isErrorObj(evald) {
			return []object.Object{evald}
		}
		res = append(res, evald)
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
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// allows us use identifiers to access builtin functions
	if builtin, ok := builtins[node.Value]; ok {
		if node.Value == "sb" {
			return evalShowBuiltInFunctions()
		}
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

func evalShowBuiltInFunctions() object.Object {
	fmt.Println(".builtins.")
	fmt.Println(".========.")
	for cStr, fn := range builtins {
		if cStr == "sb" {
			continue
		}
		fmt.Println(cStr, ": ", fn.Desc)
	}
	fmt.Println("sb : ", "shows builtin functions and their descriptions")
	// fmt.Println("sd : ", "take`s a builtin functions and shows the description")
	fmt.Println()
	return NIL
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
	case operandTypeChecks(left.Type(), right.Type(), object.STRING_OBJ):
		return evalStringInfixExpression(left, operator, right)
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

func evalStringInfixExpression(left object.Object, operator string,
	right object.Object) object.Object {
	switch operator {
	case "+":
		return evalStringConcatenationInfixExpression(left, operator, right)
	case "==":
		return evalStringEqualityInfixExpression(left, right)
	case "!=":
		return evalStringInequalityInfixExpression(left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringConcatenationInfixExpression(left object.Object, operator string,
	right object.Object) object.Object {
	lVal := left.(*object.String).Value
	rVal := right.(*object.String).Value
	return &object.String{Value: lVal + rVal}
}

func evalStringEqualityInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.String).Value
	rval := right.(*object.String).Value
	return getBooleanObject(lval == rval)
}

func evalStringInequalityInfixExpression(left object.Object, right object.Object) object.Object {
	lval := left.(*object.String).Value
	rval := right.(*object.String).Value
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
