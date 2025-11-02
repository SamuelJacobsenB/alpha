package parser

// precedências (maior = maior prioridade)
const (
	_ int = iota
	LOWEST
	LOGICALOR  // ||
	LOGICALAND // &&
	EQUALITY   // == !=
	COMPARISON // < > <= >=
	SUM        // + -
	PRODUCT    // * / %
	PREFIX     // -X !X
	CALL       // func(...)
)

var precedences = map[string]int{
	"||": LOGICALOR,
	"&&": LOGICALAND,
	"==": EQUALITY, "!=": EQUALITY,
	"<": COMPARISON, ">": COMPARISON, "<=": COMPARISON, ">=": COMPARISON,
	"+": SUM, "-": SUM,
	"*": PRODUCT, "/": PRODUCT, "%": PRODUCT,
	"(": CALL,
}
