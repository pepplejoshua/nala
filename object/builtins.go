package object

import (
	"bufio"
	"fmt"
	"os"
)

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func argumentCountMatch(given int, expected int) bool {
	return given == expected
}

func GetBuiltinByName(name string) *BuiltIn {
	for _, def := range Builtins {
		if def.Name == name {
			return def.BuiltIn
		}
	}
	return nil
}

// define builtins
func nala_len(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *String:
		return &Integer{Value: int64(len(arg.Value))}
	case *Array:
		return &Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` is not supported, got %s", args[0].Type())
	}
}

func nala_object_type(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch args[0].(type) {
	case *String:
		return &String{Value: STRING_OBJ}
	case *Boolean:
		return &String{Value: BOOLEAN_OBJ}
	case *Integer:
		return &String{Value: INTEGER_OBJ}
	case *Nil:
		return &String{Value: NIL_OBJ}
	case *Function:
		return &String{Value: FUNCTION_OBJ}
	case *BuiltIn:
		return &String{Value: BUILTIN_OBJ}
	case *Array:
		return &String{Value: ARRAY_OBJ}
	case *HashMap:
		return &String{Value: HASHMAP_OBJ}
	default:
		return newError("object type unexpected. got %s", args[0].Type())
	}
}

func nala_first(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}

	return NIL
}

func nala_last(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	lent := len(arr.Elements)
	if lent > 0 {
		return arr.Elements[lent-1]
	}

	return NIL
}

func nala_rest(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	lent := len(arr.Elements)
	if lent > 0 {
		nElems := make([]Object, lent-1)
		copy(nElems, arr.Elements[1:lent])
		return &Array{Elements: nElems}
	}

	return NIL
}

func nala_push(args ...Object) Object {
	if !argumentCountMatch(len(args), 2) {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	if args[0].Type() != ARRAY_OBJ {
		return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
	}

	arr := args[0].(*Array)
	lent := len(arr.Elements)

	nElems := make([]Object, lent+1)
	copy(nElems, arr.Elements)
	nElems[lent] = args[1]

	return &Array{Elements: nElems}
}

func nala_puts(args ...Object) Object {
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

func nala_putl(args ...Object) Object {
	if argumentCountMatch(len(args), 0) {
		fmt.Println()
		return NIL
	}

	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return NIL
}

func nala_reads(args ...Object) Object {
	if len(args) > 1 {
		return newError("wrong number of arguments. got=%d, want at most 1", len(args))
	}

	if argumentCountMatch(len(args), 1) {
		if args[0].Type() != STRING_OBJ {
			return newError("argument to `keys` must be STRING, got %s", args[0].Type())
		}

		str := args[0].(*String).Value
		fmt.Print(str)
	}

	in := os.Stdin
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		return NIL
	}
	return &String{Value: scanner.Text()}
}

func nala_hashmap_keys(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	if args[0].Type() != HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}

	hmap := args[0].(*HashMap)

	elems := []Object{}

	for _, pair := range hmap.Pairs {
		elems = append(elems, pair.Key)
	}
	return &Array{Elements: elems}
}

func nala_hashmap_values(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	if args[0].Type() != HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}

	hmap := args[0].(*HashMap)

	elems := []Object{}

	for _, pair := range hmap.Pairs {
		elems = append(elems, pair.Value)
	}
	return &Array{Elements: elems}
}

func nala_hashmap_items(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d, want=0", len(args))
	}

	if args[0].Type() != HASHMAP_OBJ {
		return newError("argument to `keys` must be HASHMAP, got %s", args[0].Type())
	}

	hmap := args[0].(*HashMap)

	elems := []Object{}

	for _, pair := range hmap.Pairs {
		nested_elems := []Object{
			pair.Key,
			pair.Value,
		}

		elems = append(elems, &Array{Elements: nested_elems})
	}
	return &Array{Elements: elems}
}

func nala_insert(args ...Object) Object {
	if !argumentCountMatch(len(args), 3) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	switch ob := args[0].(type) {
	case *HashMap:
		hashKey, ok := args[1].(Hashable)
		if !ok {
			return newError("unusable as hash key: %s", args[1].Type())
		}

		val := args[2]
		hsh := hashKey.HashKey()
		ob.Pairs[hsh] = HashPair{Key: hashKey.(Object), Value: val}
		return nil
	case *Array:
		ind, ok := args[1].(*Integer)
		if !ok {
			return newError("Array key should be INTEGER. got %s", args[1].Type())
		}

		if len(ob.Elements) == 0 || len(ob.Elements) == int(ind.Value) {
			r := nala_push(args[0], args[2])
			if r.Type() == ERROR_OBJ {
				return r
			}
			ob.Elements = r.(*Array).Elements
		}

		if int(ind.Value) > len(ob.Elements) {
			return newError("Index is greater than indexable length of Array.")
		}

		ob.Elements[int(ind.Value)] = args[2]
		return NIL
	default:
		return newError("argument to `ins` must be HASHMAP/ARRAY, got %s", args[0].Type())
	}
}

func nala_delete(args ...Object) Object {
	if !argumentCountMatch(len(args), 2) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	switch ob := args[0].(type) {
	case *HashMap:
		hashKey, ok := args[1].(Hashable)
		if !ok {
			return newError("unusable as hash key: %s", args[1].Type())
		}

		if _, ok := ob.Pairs[hashKey.HashKey()]; !ok {
			return newError("key does not exist in HashMap")
		} else {
			delete(ob.Pairs, hashKey.HashKey())
		}
		return NIL
	case *Array:
		ind, ok := args[1].(*Integer)
		if !ok {
			return newError("Array key should be INTEGER. got %s", args[1].Type())
		}
		if int(ind.Value) > len(ob.Elements) {
			return newError("Index is greater than indexable length of Array.")
		}

		in := int(ind.Value)
		elems := make([]Object, 0)
		elems = append(elems, ob.Elements[:in]...)
		elems = append(elems, ob.Elements[in+1:]...)
		ob.Elements = elems
	default:
		return newError("argument to `del` must be HASHMAP/ARRAY, got %s", args[0].Type())
	}
	return NIL
}

func nala_copy(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	switch obj := args[0].(type) {
	case *Array:
		lent := len(obj.Elements)
		if lent > 0 {
			nCopy := make([]Object, lent)
			copy(nCopy, obj.Elements)
			return &Array{Elements: nCopy}
		}
		return &Array{Elements: obj.Elements}
	case *HashMap:
		pairs := make(map[HashKey]HashPair)

		for k, pair := range obj.Pairs {
			key := pair.Key
			val := pair.Value
			pairs[k] = HashPair{Key: key, Value: val}
		}
		return &HashMap{Pairs: pairs}
	default:
		return newError("argument to `copy` is not supported, got %s", obj.Type())
	}
}

func nala_showbuiltin_desc(args ...Object) Object {
	if !argumentCountMatch(len(args), 1) {
		return newError("wrong number of arguments. got=%d", len(args))
	}

	if args[0].Type() == BUILTIN_OBJ {
		fmt.Println(args[0].Inspect())
	}
	return NIL
}

var Builtins = []struct {
	Name    string
	BuiltIn *BuiltIn
}{
	{
		Name:    "len",
		BuiltIn: &BuiltIn{Fn: nala_len, Desc: "calculates the length of a Nala iterable"},
	},
	{
		Name:    "type",
		BuiltIn: &BuiltIn{Fn: nala_object_type, Desc: "shows the type of a Nala object"},
	},
	{
		Name:    "first",
		BuiltIn: &BuiltIn{Fn: nala_first, Desc: "returns the first element of an Array"},
	},
	{
		Name:    "last",
		BuiltIn: &BuiltIn{Fn: nala_last, Desc: "returns the last element of an Array"},
	},
	{
		Name: "rest",
		BuiltIn: &BuiltIn{
			Fn:   nala_rest,
			Desc: "returns a new copy of passed Array excluding first element"},
	},
	{
		Name:    "push",
		BuiltIn: &BuiltIn{Fn: nala_push, Desc: "pushes a new element to the back of an Array"},
	},
	{
		Name:    "puts",
		BuiltIn: &BuiltIn{Fn: nala_puts, Desc: "prints to standard output on the same line, with a terminating newline.\nTakes 0 or more arguments"},
	},
	{
		Name:    "putl",
		BuiltIn: &BuiltIn{Fn: nala_putl, Desc: "prints to standard output on multiple lines.\nTakes 0 or more arguments"},
	},
	{
		Name:    "reads",
		BuiltIn: &BuiltIn{Fn: nala_reads, Desc: "reads string from standard input. Takes conditional prompt string"},
	},
	{
		Name:    "keys",
		BuiltIn: &BuiltIn{Fn: nala_hashmap_keys, Desc: "returns the keys of a HashMap in an Array"},
	},
	{
		Name:    "values",
		BuiltIn: &BuiltIn{Fn: nala_hashmap_values, Desc: "returns the keys of a HashMap in an Array"},
	},
	{
		Name:    "items",
		BuiltIn: &BuiltIn{Fn: nala_hashmap_items, Desc: "returns an Array of Arrays containing Key, Value of a HashMap."},
	},
	{
		Name:    "ins",
		BuiltIn: &BuiltIn{Fn: nala_insert, Desc: "inserts a Value at a Key/Index in a HashMap/Array"},
	},
	{
		Name:    "del",
		BuiltIn: &BuiltIn{Fn: nala_delete, Desc: "delete the Value at a Key/Index from a HashMap/Array"},
	},
	{
		Name:    "copy",
		BuiltIn: &BuiltIn{Fn: nala_copy, Desc: "returns a copy of an Array or HashMap"},
	},
	{
		Name:    "desc",
		BuiltIn: &BuiltIn{Fn: nala_showbuiltin_desc, Desc: "takes a builtin functions and shows the description"},
	},
}
