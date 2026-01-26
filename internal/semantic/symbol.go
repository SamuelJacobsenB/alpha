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
	KindImport
)

type Symbol struct {
	Name string
	Kind SymbolKind
	Type Type // Alterado para semantic.Type
	Node parser.Node
}
