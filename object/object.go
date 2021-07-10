package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"nala/ast"
	"nala/opcode"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ           = "INTEGER"
	BOOLEAN_OBJ           = "BOOLEAN"
	NIL_OBJ               = "NIL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING_OBJ            = "STRING"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	HASHMAP_OBJ           = "HASHMAP"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNC"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Value       int64
	HashableKey *HashKey
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) HashKey() HashKey {
	if i.HashableKey == nil {
		i.HashableKey = &HashKey{Type: i.Type(), HashValue: uint64(i.Value)}
	}
	return *i.HashableKey
}

type Boolean struct {
	Value       bool
	HashableKey *HashKey
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	if b.HashableKey == nil {
		var val uint64

		if b.Value {
			val = 1
		} else {
			val = 0
		}
		b.HashableKey = &HashKey{Type: b.Type(), HashValue: val}
	}
	return *b.HashableKey
}

type Nil struct{}

func (n *Nil) Type() ObjectType { return NIL_OBJ }
func (n *Nil) Inspect() string  { return "nil" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "Error: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn (")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type CompiledFunction struct {
	Instructions opcode.Instructions
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type String struct {
	Value       string
	HashableKey *HashKey
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	if s.HashableKey == nil {
		h := fnv.New64a()
		h.Write([]byte(s.Value))
		s.HashableKey = &HashKey{Type: s.Type(), HashValue: h.Sum64()}
	}
	return *s.HashableKey
}

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elems := []string{}
	for _, p := range a.Elements {
		elems = append(elems, p.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elems, ", "))
	out.WriteString("]")

	return out.String()
}

// the key used in our HashMaps
// hashed from true Values of Expressions
// to prevent pointer comparison
// TODO: cache these values so they are not recomputed everytime
type HashKey struct {
	Type      ObjectType
	HashValue uint64
}

// the Value stored in a HashMap
// contains true key:value passed by user
type HashPair struct {
	Key   Object
	Value Object
}

// HashMap
type HashMap struct {
	Pairs map[HashKey]HashPair
}

func (hm *HashMap) Type() ObjectType { return HASHMAP_OBJ }
func (hm *HashMap) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}

	for _, p := range hm.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			p.Key.Inspect(), p.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type BuiltInFunction func(args ...Object) Object

type BuiltIn struct {
	Fn   BuiltInFunction
	Desc string
}

func (b *BuiltIn) Type() ObjectType { return BUILTIN_OBJ }
func (b *BuiltIn) Inspect() string  { return fmt.Sprintf("builtin function: %q", b.Desc) }

// func () Type() ObjectType { return }
// func () Inspect() string { return }
// func () HashKey() HashKey { return }

// Environment and Binding
type NameObjectPairs map[string]Object

type Environment struct {
	store   NameObjectPairs
	extends *Environment
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.extends = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(NameObjectPairs)
	return &Environment{store: s, extends: nil}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.extends != nil {
		obj, ok = e.extends.Get(name)
	}
	return obj, ok
}

func (e *Environment) GetStore() NameObjectPairs { return e.store }

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
