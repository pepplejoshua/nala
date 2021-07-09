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
