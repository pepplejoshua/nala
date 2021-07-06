package evaluator

import (
	"fmt"
	"nala/ast"
	"nala/object"
	"nala/token"
)

func DefineMacros(prog *ast.Program, env *object.Environment) {
	indexes := []int{}

	for i, stmt := range prog.Statements {
		if isMacroDefinition(stmt) {
			addMacro(stmt, env)
			indexes = append(indexes, i)
		}
	}

	// for i := len(indexes) - 1; i >= 0; i-- {
	// 	defIn := indexes[i]
	// remove the macro statements from the bunch of executable statements
	// if I don't remove them, can I use them for future references to the Macro?
	// prog.Statements = append(prog.Statements[:defIn],
	// prog.Statements[defIn+1:]...)
	// }
}

// func RunMacro(prog ast.Node, env *object.Environment) ast.Node {
// 	modifier := func(node ast.Node) ast.Node {
// 		mlit, ok := node.(*ast.MacroLiteral)
// 		if !ok {
// 			return node
// 		}

// 		macro, ok := isMacroCall(call, env)
// 		if !ok {
// 			return node
// 		}

// 		args := quoteArgs(call)
// 		macroEnv := extendMacroEnv(macro, args)

// 		eval := Eval(macro.Body, macroEnv)

// 		quote, ok := eval.(*object.Quote)
// 		if !ok {
// 			panic("we only support return quoted results from macros (AST-nodes only)")
// 		}
// 		return quote.CodeNode
// 	}
// 	return ast.Modify(prog, modifier)
// }

// equivalent to calling the Macro when it is encountered
func ExpandMacros(prog ast.Node, env *object.Environment) ast.Node {
	modifier := func(node ast.Node) ast.Node {
		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		macro, ok := isMacroCall(call, env)
		if !ok {
			return node
		}

		args := quoteArgs(call)
		macroEnv := extendMacroEnv(macro, args)

		eval := Eval(macro.Body, macroEnv)

		quote, ok := eval.(*object.Quote)
		if !ok {
			panic("we only support return quoted results from macros (AST-nodes only)")
		}
		return quote.CodeNode
	}
	return ast.Modify(prog, modifier)
}

func isMacroCall(call *ast.CallExpression,
	env *object.Environment) (*object.Macro, bool) {
	id, ok := call.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(id.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

func quoteArgs(call *ast.CallExpression) []*object.Quote {
	args := []*object.Quote{}

	for _, a := range call.Arguments {
		args = append(args, &object.Quote{CodeNode: a})
	}
	return args
}

func extendMacroEnv(macro *object.Macro,
	args []*object.Quote) *object.Environment {
	ext := object.NewEnclosedEnvironment(macro.Env)

	for i, p := range macro.Parameters {
		ext.Set(p.Value, args[i])
	}

	return ext
}

func addMacro(node ast.Statement, env *object.Environment) {
	let, _ := node.(*ast.LetStatement)
	macLiteral, _ := let.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters:   macLiteral.Parameters,
		Env:          env,
		Body:         macLiteral.Body,
		MacroLiteral: macLiteral,
	}

	env.Set(let.Name.Value, macro)
}

// this function strictly enforces what a Macro definition is
// This is valid:
// let mac = macro(x) { quote(x) }
// let invalid = mac
// the above line is invalid
func isMacroDefinition(node ast.Statement) bool {
	let, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, ok = let.Value.(*ast.MacroLiteral)
	return ok
}

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
