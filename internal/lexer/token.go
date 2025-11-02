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
	"int": {}, "var": {}, "const": {}, "function": {},
	"class": {}, "if": {}, "else": {}, "while": {}, "for": {},
	"return": {}, "true": {}, "false": {}, "null": {}, "type": {},
	"enum": {}, "import": {}, "export": {}, "constructor": {},
	"break": {}, "continue": {}, "switch": {}, "case": {}, "default": {}, "package": {},
}
