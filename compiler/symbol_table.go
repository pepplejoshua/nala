package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltInScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer          *SymbolTable // enclosing symbol table
	store          map[string]Symbol
	numDefinitions int
}

func (st *SymbolTable) Define(id string) Symbol {
	// can check if symbol actually already exists and reuse it's index.
	// but not sure of the implications of that just yet
	existing, ok := st.store[id]
	if ok {
		return existing
	} else {
		sym := Symbol{
			Name:  id,
			Index: st.numDefinitions,
		}

		if st.Outer == nil {
			sym.Scope = GlobalScope
		} else {
			sym.Scope = LocalScope
		}

		st.store[id] = sym
		st.numDefinitions++
		return sym
	}
}

func (st *SymbolTable) Resolve(id string) (Symbol, bool) {
	existing, ok := st.store[id]
	if !ok && st.Outer != nil {
		return st.Outer.Resolve(id)
	}
	return existing, ok
}

func (st *SymbolTable) DefineBuiltin(index int, id string) Symbol {
	sym := Symbol{
		Name:  id,
		Scope: BuiltInScope,
		Index: index,
	}
	st.store[id] = sym
	return sym
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{
		store:          s,
		numDefinitions: 0,
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}
