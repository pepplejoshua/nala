package evaluator

import (
	"nala/ast"
	"nala/object"
)

var (
	NIL   = &object.Nil{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return getBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(left, node.Operator, right)
	}

	return nil
}

func evalStatements(stmts []ast.Statement) object.Object {
	var res object.Object

	for _, stmt := range stmts {
		res = Eval(stmt)
	}

	return res
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangPrefixExpression(right)
	case "-":
		return evalMinusPrefixExpression(right)
	default:
		return NIL
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
		return NIL
	}

	val := right.(*object.Integer).Value
	return &object.Integer{Value: -val}
}

func evalInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch operator {
	case "+":
		return evalAdditionInfixExpression(left, right)
	case "-":
		return evalMinusInfixExpression(left, right)
	case "*":
		return evalMultiplicationInfixExpression(left, right)
	case "/":
		return evalDivisionInfixExpression(left, right)
	case "%":
		return evalModuloInfixExpression(left, right)
	default:
		return NIL
	}

}

func evalAdditionInfixExpression(left object.Object, right object.Object) object.Object {
	if !operandTypeChecks(left.Type(), right.Type(), object.INTEGER_OBJ) {
		return NIL
	}

	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return &object.Integer{Value: lval + rval}
}

func evalMinusInfixExpression(left object.Object, right object.Object) object.Object {
	if !operandTypeChecks(left.Type(), right.Type(), object.INTEGER_OBJ) {
		return NIL
	}

	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return &object.Integer{Value: lval - rval}
}

func evalMultiplicationInfixExpression(left object.Object, right object.Object) object.Object {
	if !operandTypeChecks(left.Type(), right.Type(), object.INTEGER_OBJ) {
		return NIL
	}

	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value
	return &object.Integer{Value: lval * rval}
}

func evalDivisionInfixExpression(left object.Object, right object.Object) object.Object {
	if !operandTypeChecks(left.Type(), right.Type(), object.INTEGER_OBJ) {
		return NIL
	}

	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value

	if rval == 0 {
		return NIL // ZeroDivisionError
	}
	return &object.Integer{Value: lval / rval}
}

func evalModuloInfixExpression(left object.Object, right object.Object) object.Object {
	if !operandTypeChecks(left.Type(), right.Type(), object.INTEGER_OBJ) {
		return NIL
	}

	lval := left.(*object.Integer).Value
	rval := right.(*object.Integer).Value

	if rval == 0 {
		return NIL // ZeroDivisionError
	}
	return &object.Integer{Value: lval % rval}
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
