package semantic

import "github.com/alpha/internal/parser"

type SymbolKind int

const (
	KindVar SymbolKind = iota
	KindConst
	KindFunction
	KindStruct
	KindTypeAlias
	KindGenericParam
)

type Symbol struct {
	Name string
	Kind SymbolKind
	Type parser.Type // O tipo declarado (AST Node)
	Node parser.Node // Referência ao nó da declaração (para erros)
}
