package evaluator

import (
	"bufio"
	"fmt"
	"nala/object"
	"os"
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
	case *object.HashMap:
		return &object.String{Value: object.HASHMAP_OBJ}
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

func nala_puts(args ...object.Object) object.Object {
	if argumentCountMatch(len(args), 0) {
		fmt.Print()
		return NIL
	}

	for _, arg := range args {
		fmt.Print(arg.Inspect())
	}
	fmt.Println()
	return NIL
}

func nala_putl(args ...object.Object) object.Object {
	if argumentCountMatch(len(args), 0) {
		fmt.Println()
		return NIL
	}

	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return NIL
}

func nala_reads(args ...object.Object) object.Object {
	if len(args) > 1 {
		return newError("wrong number of arguments. got=%d, want at most 1", len(args))
	}

	if argumentCountMatch(len(args), 1) {
		if args[0].Type() != object.STRING_OBJ {
			return newError("argument to `keys` must be STRING, got %s", args[0].Type())
		}

		str := args[0].(*object.String).Value
		fmt.Print(str)
	}

	in := os.Stdin
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		return NIL
	}
	return &object.String{Value: scanner.Text()}
}

func nala_hashmap_keys(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	if args[0].Type() != object.HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}

	hmap := args[0].(*object.HashMap)

	elems := []object.Object{}

	for _, pair := range hmap.Pairs {
		elems = append(elems, pair.Key)
	}
	return &object.Array{Elements: elems}
}

func nala_hashmap_values(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	if args[0].Type() != object.HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}

	hmap := args[0].(*object.HashMap)

	elems := []object.Object{}

	for _, pair := range hmap.Pairs {
		elems = append(elems, pair.Value)
	}
	return &object.Array{Elements: elems}
}

func nala_hashmap_items(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	if args[0].Type() != object.HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}

	hmap := args[0].(*object.HashMap)

	elems := []object.Object{}

	for _, pair := range hmap.Pairs {
		nested_elems := []object.Object{
			pair.Key,
			pair.Value,
		}

		elems = append(elems, &object.Array{Elements: nested_elems})
	}
	return &object.Array{Elements: elems}
}

// export builtins to REPL
var builtins = MapofIDtoBuiltin{
	"len":    &object.BuiltIn{Fn: nala_len},
	"type":   &object.BuiltIn{Fn: nala_object_type},
	"first":  &object.BuiltIn{Fn: nala_first},
	"last":   &object.BuiltIn{Fn: nala_last},
	"rest":   &object.BuiltIn{Fn: nala_rest},
	"push":   &object.BuiltIn{Fn: nala_push},
	"puts":   &object.BuiltIn{Fn: nala_puts},
	"putl":   &object.BuiltIn{Fn: nala_putl},
	"reads":  &object.BuiltIn{Fn: nala_reads},
	"keys":   &object.BuiltIn{Fn: nala_hashmap_keys},
	"values": &object.BuiltIn{Fn: nala_hashmap_values},
	"items":  &object.BuiltIn{Fn: nala_hashmap_items},
	// "loadf":  &object.BuiltIn{Fn: nala_loadf},
}
