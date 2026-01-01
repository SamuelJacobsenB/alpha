package semantic

type Scope struct {
	Outer   *Scope             // Escopo pai (nil se for global)
	Symbols map[string]*Symbol // Mapa de símbolos neste escopo
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		Outer:   outer,
		Symbols: make(map[string]*Symbol),
	}
}

// Define insere um símbolo no escopo ATUAL
func (s *Scope) Define(name string, sym *Symbol) bool {
	if _, exists := s.Symbols[name]; exists {
		return false // Erro: Já declarado neste escopo
	}
	s.Symbols[name] = sym
	return true
}

// Resolve busca recursivamente nos escopos pais
func (s *Scope) Resolve(name string) *Symbol {
	if sym, ok := s.Symbols[name]; ok {
		return sym
	}
	if s.Outer != nil {
		return s.Outer.Resolve(name)
	}
	return nil
}
