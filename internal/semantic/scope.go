package semantic

// Scope simples encadeado usado pelo checker
type Scope struct {
	parent *Scope
	table  map[string]*Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent: parent, table: make(map[string]*Symbol)}
}

// Define registra um símbolo no escopo atual.
// Retorna erro se já existir definição local com mesmo nome.
func (s *Scope) Define(sym *Symbol) error {
	if _, ok := s.table[sym.Name]; ok {
		return &SemanticErr{Msg: "symbol already defined in this scope", Line: sym.DeclLine, Col: sym.DeclCol}
	}
	s.table[sym.Name] = sym
	return nil
}

// Resolve busca um símbolo no escopo atual e em ancestrais.
// Retorna (sym, true) se encontrado, (nil, false) caso contrário.
func (s *Scope) Resolve(name string) (*Symbol, bool) {
	if sym, ok := s.table[name]; ok {
		return sym, true
	}
	if s.parent != nil {
		return s.parent.Resolve(name)
	}
	return nil, false
}

// HasLocal verifica se o nome existe apenas no escopo atual.
func (s *Scope) HasLocal(name string) bool {
	_, ok := s.table[name]
	return ok
}
