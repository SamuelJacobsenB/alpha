package semantic

import "fmt"

// SemanticErr é usado internamente para sinalizar erros com posição
type SemanticErr struct {
	Msg  string
	Line int
	Col  int
}

func (e *SemanticErr) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("semantic: %s at %d:%d", e.Msg, e.Line, e.Col)
	}
	return fmt.Sprintf("semantic: %s", e.Msg)
}
