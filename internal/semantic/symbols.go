package semantic

// Symbol representa uma entrada na tabela de símbolos
type Symbol struct {
	Name     string
	Type     Type
	Mutable  bool // var = true, const = false
	DeclLine int
	DeclCol  int
}
