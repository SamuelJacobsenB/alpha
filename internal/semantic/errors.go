package semantic

import (
	"fmt"
)

type SemanticError struct {
	Msg  string
	Line int
	Col  int
}

func (e SemanticError) Error() string {
	return fmt.Sprintf("[Semantic Error] @ %d:%d: %s", e.Line, e.Col, e.Msg)
}

// Interface para nós que têm posição (já definida no seu parser)
type NodeWithPos interface {
	// Assumindo que você pode extrair linha/coluna do parser.Node
	// Se o nodePos() for privado, você precisará de getters no parser
	// ou passar o token correspondente.
}
