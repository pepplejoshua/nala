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
	case *object.Quote:
		return &object.String{Value: object.QUOTE_OBJ}
	case *object.Macro:
		return &object.String{Value: object.MACRO_OBJ}
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

func nala_hashmap_insert(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 3) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	if args[0].Type() != object.HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}
	hmap := args[0].(*object.HashMap)
	hashKey, ok := args[1].(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", args[1].Type())
	}

	val := args[2]
	hsh := hashKey.HashKey()
	hmap.Pairs[hsh] = object.HashPair{Key: hashKey.(object.Object), Value: val}
	return NIL
}

func nala_copy(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	switch obj := args[0].(type) {
	case *object.Array:
		lent := len(obj.Elements)
		if lent > 0 {
			nCopy := make([]object.Object, lent)
			copy(nCopy, obj.Elements)
			return &object.Array{Elements: nCopy}
		}
		return &object.Array{Elements: obj.Elements}
	case *object.HashMap:
		pairs := make(map[object.HashKey]object.HashPair)

		for k, pair := range obj.Pairs {
			key := pair.Key
			val := pair.Value
			pairs[k] = object.HashPair{Key: key, Value: val}
		}
		return &object.HashMap{Pairs: pairs}
	default:
		return newError("argument to `copy` is not supported, got %s", obj.Type())
	}
}

func nala_showbuiltin_info(args ...object.Object) object.Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	if args[0].Type() == object.BUILTIN_OBJ {
		fmt.Println(args[0].Inspect())
	}
	return NIL
}

func nala_showuserdef_fns(args ...object.Object) object.Object {

	return NIL
}

// TODOs:
// Iterable interface: Array, String, Vector

// export builtins to REPL
var builtins = MapofIDtoBuiltin{
	"len":   &object.BuiltIn{Fn: nala_len, Desc: "calculates the length of a Nala iterable"},
	"type":  &object.BuiltIn{Fn: nala_object_type, Desc: "shows the type of a Nala object"},
	"first": &object.BuiltIn{Fn: nala_first, Desc: "returns the first element of an Array"},
	"last":  &object.BuiltIn{Fn: nala_last, Desc: "returns the last element of an Array"},
	"rest": &object.BuiltIn{
		Fn:   nala_rest,
		Desc: "returns a new copy of passed Array excluding first element"},
	"push":   &object.BuiltIn{Fn: nala_push, Desc: "pushes a new element to the back of an Array"},
	"puts":   &object.BuiltIn{Fn: nala_puts, Desc: "prints to standard output on the same line, with a terminating newline.\nTakes 0 or more arguments"},
	"putl":   &object.BuiltIn{Fn: nala_putl, Desc: "prints to standard output on multiple lines.\nTakes 0 or more arguments"},
	"reads":  &object.BuiltIn{Fn: nala_reads, Desc: "reads string from standard input. Takes conditional prompt string"},
	"keys":   &object.BuiltIn{Fn: nala_hashmap_keys, Desc: "returns the keys of a HashMap in an Array"},
	"values": &object.BuiltIn{Fn: nala_hashmap_values, Desc: "returns the keys of a HashMap in an Array"},
	"items":  &object.BuiltIn{Fn: nala_hashmap_items, Desc: "returns an Array of Arrays containing Key, Value of a HashMap."},
	"ins":    &object.BuiltIn{Fn: nala_hashmap_insert, Desc: "inserts a Value at a Key in a HashMap"},
	"copy":   &object.BuiltIn{Fn: nala_copy, Desc: "returns a copy of an Array or HashMap"},
	"sb":     &object.BuiltIn{Fn: nil},
	"sd":     &object.BuiltIn{Fn: nala_showbuiltin_info, Desc: "takes a builtin functions and shows the description"},
	"sf":     &object.BuiltIn{Fn: nala_showuserdef_fns, Desc: "shows all user bound functions in environment"},
	// "unquote": &object.BuiltIn{Fn: nala_outer_unquote, Desc: "used as an external unquote for the Quote objects returned by Macros or Fns."},
	// "loadf":  &object.BuiltIn{Fn: nala_loadf},
}
