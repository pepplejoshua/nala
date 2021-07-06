package evaluator

import (
	"fmt"
	"nala/ast"
	"nala/object"
	"nala/token"
)

func quote(node ast.Node, env *object.Environment) object.Object {
	node = evalUnquotedCalls(node, env)
	return &object.Quote{CodeNode: node}
}

func evalUnquotedCalls(quoted ast.Node, env *object.Environment) ast.Node {
	modifier := func(node ast.Node) ast.Node {
		if !isUnquoteCall(node) {
			return node
		}

		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		if len(call.Arguments) != 1 {
			return node
		}

		unquoted := Eval(call.Arguments[0], env)
		return convertObjectToASTNode(unquoted)
	}
	return ast.Modify(quoted, modifier)
}

func isUnquoteCall(node ast.Node) bool {
	call, ok := node.(*ast.CallExpression)
	if !ok {
		return false
	}
	return call.Function.TokenLiteral() == "unquote"
}

func convertObjectToASTNode(obj object.Object) ast.Node {
	switch obj := obj.(type) {
	case *object.Integer:
		tok := makeToken(token.INT, fsp(obj.Value))
		return &ast.IntegerLiteral{Token: tok, Value: obj.Value}
	case *object.Boolean:
		var t token.Token
		if obj.Value {
			t = makeToken(token.TRUE, "true")
		} else {
			t = makeToken(token.FALSE, "false")
		}
		return &ast.Boolean{Token: t, Value: obj.Value}
	case *object.String:
		t := makeToken(token.STRING, obj.Value)
		return &ast.StringLiteral{Token: t, Value: obj.Value}
	case *object.Quote:
		return obj.CodeNode
	default:
		return nil
	}
}

func makeToken(ttype token.TokenType, literal string) token.Token {
	return token.Token{Type: ttype, Literal: literal}
}

func fsp(i int64) string { return fmt.Sprintf("%d", i) }
