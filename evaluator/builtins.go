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
	case *object.Array:
		return &object.String{Value: object.ARRAY_OBJ}
	default:
		return newError("object type unexpected. got %s", args[0].Type())
	}
}

func nala_first(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return NIL
}

func nala_last(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	lent := len(arr.Elements)
	if lent > 0 {
		return arr.Elements[lent-1]
	}

	return NIL
}

func nala_rest(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	lent := len(arr.Elements)
	if lent > 0 {
		nElems := make([]object.Object, lent-1)
		copy(nElems, arr.Elements[1:lent])
		return &object.Array{Elements: nElems}
	}

	return NIL
}

func nala_push(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 2) {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	if args[0].Type() != object.ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*object.Array)
	lent := len(arr.Elements)

	nElems := make([]object.Object, lent+1)
	copy(nElems, arr.Elements)
	nElems[lent] = args[1]

	return &object.Array{Elements: nElems}
}

// export builtins to REPL
var builtins = MapofIDtoBuiltin{
	"len":   &object.BuiltIn{Fn: nala_len},
	"type":  &object.BuiltIn{Fn: nala_object_type},
	"first": &object.BuiltIn{Fn: nala_first},
	"last":  &object.BuiltIn{Fn: nala_last},
	"rest":  &object.BuiltIn{Fn: nala_rest},
	"push":  &object.BuiltIn{Fn: nala_push},
}
