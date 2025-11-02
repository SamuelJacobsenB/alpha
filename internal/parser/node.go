package parser

// Node é a interface comum a todos os nós da AST.
// Serve para permitir anotações (ex: map[parser.Node]Type) no semantic/checker.
type Node interface {
	nodePos()
}
