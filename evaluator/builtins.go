package evaluator

import (
	"nala/object"
)

type MapofIDtoBuiltin map[string]*object.BuiltIn

func argumentCountMatch(given int, expected int) bool {
	return given == expected
}

// define builtins
func nala_len(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` is not supported, got %s", args[0].Type())
	}
}

func nala_object_type(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch args[0].(type) {
	case *object.String:
		return &object.String{Value: object.STRING_OBJ}
	case *object.Boolean:
		return &object.String{Value: object.BOOLEAN_OBJ}
	case *object.Integer:
		return &object.String{Value: object.INTEGER_OBJ}
	case *object.Nil:
		return &object.String{Value: object.NIL_OBJ}
	case *object.Function:
		return &object.String{Value: object.FUNCTION_OBJ}
	case *object.BuiltIn:
		return &object.String{Value: object.BUILTIN_OBJ}
	default:
		return newError("object type unexpected. got %s", args[0].Type())
	}
}

// export builtins to REPL
var builtins = MapofIDtoBuiltin{
	"len":  &object.BuiltIn{Fn: nala_len},
	"type": &object.BuiltIn{Fn: nala_object_type},
}
