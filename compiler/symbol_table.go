package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltInScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
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
	FreeSymbols    []Symbol
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
		existing, ok = st.Outer.Resolve(id)
		if !ok {
			return existing, ok
		}

		if existing.Scope == GlobalScope || existing.Scope == BuiltInScope {
			return existing, ok
		}

		free := st.defineFree(existing)
		return free, true
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

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	sym := Symbol{
		Name:  original.Name,
		Scope: FreeScope,
		Index: len(st.FreeSymbols) - 1,
	}

	st.store[original.Name] = sym
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
