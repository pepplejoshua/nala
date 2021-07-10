package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func (st *SymbolTable) Define(id string) Symbol {
	// can check if symbol actually already exists and reuse it's index.
	// but not sure of the implications of that just yet
	existing, ok := st.Resolve(id)
	if ok {
		return existing
	} else {
		sym := Symbol{
			Name:  id,
			Scope: GlobalScope,
			Index: st.numDefinitions,
		}

		st.store[id] = sym
		st.numDefinitions++
		return sym
	}
}

func (st *SymbolTable) Resolve(id string) (Symbol, bool) {
	existing, ok := st.store[id]
	return existing, ok
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{
		store:          s,
		numDefinitions: 0,
	}
}
