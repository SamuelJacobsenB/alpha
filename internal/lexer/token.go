package lexer

// TokenType enumera tipos de token
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
	// Tipos primitivos
	"int": {}, "string": {}, "float": {}, "bool": {}, "void": {},
	"byte": {}, "char": {}, "double": {}, "error": {}, "component": {},

	// Palavras-chave de declaração
	"var": {}, "const": {}, "function": {}, "class": {}, "type": {}, "enum": {},

	// Controle de fluxo
	"if": {}, "else": {}, "while": {}, "do": {}, "for": {}, "in": {}, "return": {},
	"break": {}, "continue": {}, "switch": {}, "case": {}, "default": {},

	// Valores
	"true": {}, "false": {}, "null": {},

	// Módulos
	"import": {}, "export": {}, "package": {}, "from": {}, "as": {},

	// OOP
	"constructor": {}, "method": {}, "this": {}, "new": {},

	// Outros
	"typeof": {},
}
