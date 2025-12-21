package lexer

type TokenType int

const (
	EOF TokenType = iota
	ERROR
	KEYWORD
	IDENT
	INT
	FLOAT
	STRING
	OP
	GENERIC // T, U, etc. (parâmetros genéricos)
)

type Token struct {
	Type   TokenType
	Lexeme string // texto bruto
	Value  string // valor normalizado (p.ex. string sem aspas)
	Line   int
	Col    int
}

var keywords = map[string]struct{}{
	"int": {}, "string": {}, "float": {}, "bool": {}, "void": {},
	"byte": {}, "char": {}, "double": {}, "error": {}, "component": {},
	"var": {}, "const": {}, "function": {}, "class": {}, "type": {}, "enum": {},
	"if": {}, "else": {}, "while": {}, "do": {}, "for": {}, "in": {}, "return": {},
	"break": {}, "continue": {}, "switch": {}, "case": {}, "default": {},
	"true": {}, "false": {}, "null": {},
	"import": {}, "export": {}, "package": {}, "from": {}, "as": {},
	"constructor": {}, "method": {}, "this": {}, "new": {},
	"typeof": {}, "generic": {},
}
