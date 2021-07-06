package evaluator

import (
	"nala/ast"
	"nala/object"
)

func quote(node ast.Node) object.Object {
	return &object.Quote{CodeNode: node}
}

func unquote(quoted *object.Quote, env *object.Environment) object.Object {
	code := quoted.CodeNode

	eval := Eval(code, env)
	return eval
}
