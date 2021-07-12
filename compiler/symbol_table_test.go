package compiler

import "testing"

type SymbolTableTest map[string]Symbol

func TestDefine(t *testing.T) {
	expected := SymbolTableTest{
		"a": Symbol{
			Name:  "a",
			Scope: GlobalScope,
			Index: 0,
		},
		"b": Symbol{
			Name:  "b",
			Scope: GlobalScope,
			Index: 1,
		},
		"c": {
			Name:  "c",
			Scope: LocalScope,
			Index: 0,
		},
		"d": {
			Name:  "d",
			Scope: LocalScope,
			Index: 1,
		},
		"e": {
			Name:  "e",
			Scope: LocalScope,
			Index: 0,
		},
		"f": {
			Name:  "f",
			Scope: LocalScope,
			Index: 1,
		},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}

	first := NewEnclosedSymbolTable(global)

	c := first.Define("c")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}

	d := first.Define("d")
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}

	second := NewEnclosedSymbolTable(first)
	e := second.Define("e")
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}

	f := second.Define("f")
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}

}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		{
			Name:  "a",
			Scope: GlobalScope,
			Index: 0,
		},
		{
			Name:  "b",
			Scope: GlobalScope,
			Index: 1,
		},
	}

	for _, sym := range expected {
		res, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if res != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, res)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{
			Name:  "a",
			Scope: GlobalScope,
			Index: 0,
		},
		{
			Name:  "b",
			Scope: GlobalScope,
			Index: 1,
		},
		{
			Name:  "c",
			Scope: LocalScope,
			Index: 0,
		},
		{
			Name:  "d",
			Scope: LocalScope,
			Index: 1,
		},
	}

	for _, sym := range expected {
		res, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if res != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, res)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	first := NewEnclosedSymbolTable(global)
	first.Define("c")
	first.Define("d")

	second := NewEnclosedSymbolTable(first)
	second.Define("e")
	second.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			table: first,
			expectedSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			table: second,
			expectedSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			res, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s is not resolvable", sym.Name)
				continue
			}

			if res != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, res)
			}
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	glob := NewSymbolTable()
	first := NewEnclosedSymbolTable(glob)
	sec := NewEnclosedSymbolTable(first)

	expected := []Symbol{
		{
			Name:  "a",
			Scope: BuiltInScope,
			Index: 0,
		},
		{
			Name:  "c",
			Scope: BuiltInScope,
			Index: 1,
		},
		{
			Name:  "e",
			Scope: BuiltInScope,
			Index: 2,
		},
		{
			Name:  "f",
			Scope: BuiltInScope,
			Index: 3,
		},
	}

	for i, v := range expected {
		glob.DefineBuiltin(i, v.Name)
	}

	// make sure definition in Builtin scope is visible from every scope
	for _, table := range []*SymbolTable{glob, first, sec} {
		for _, sym := range expected {
			res, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}

			if res != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, res)
			}
		}
	}
}
