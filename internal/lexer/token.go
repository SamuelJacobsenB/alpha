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
)

// Token representa um token emitido pelo lexer
type Token struct {
	Type   TokenType
	Lexeme string // texto bruto
	Value  string // valor normalizado (p.ex. string sem aspas)
	Line   int
	Col    int
}

// keywords simples
var keywords = map[string]struct{}{
	// Tipos primitivos
	"int": {}, "string": {}, "float": {}, "boolean": {}, "void": {},
	"byte": {}, "char": {}, "double": {}, "error": {}, "component": {},

	// Palavras-chave de declaração
	"var": {}, "const": {}, "function": {}, "class": {}, "type": {}, "enum": {},

	// Controle de fluxo
	"if": {}, "else": {}, "while": {}, "for": {}, "in": {}, "return": {},
	"break": {}, "continue": {}, "switch": {}, "case": {}, "default": {},

	// Valores
	"true": {}, "false": {}, "null": {},

	// Módulos
	"import": {}, "export": {}, "package": {},

	// OOP
	"constructor": {},
}
