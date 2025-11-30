package parser

const (
	_ int = iota
	LOWEST
	ASSIGNMENT // = += -= etc.
	LOGICALOR  // ||
	LOGICALAND // &&
	EQUALITY   // == !=
	COMPARISON // < > <= >=
	SUM        // + -
	PRODUCT    // * / %
	PREFIX     // -X !X ++X --X
	CALL       // func(...)
	INDEX      // [] - acesso a array
	POSTFIX    // X++ X--
)

var precedences = map[string]int{
	"=":  ASSIGNMENT,
	"+=": ASSIGNMENT,
	"-=": ASSIGNMENT,
	"*=": ASSIGNMENT,
	"/=": ASSIGNMENT,
	"||": LOGICALOR,
	"&&": LOGICALAND,
	"==": EQUALITY,
	"!=": EQUALITY,
	"<":  COMPARISON,
	">":  COMPARISON,
	"<=": COMPARISON,
	">=": COMPARISON,
	"+":  SUM,
	"-":  SUM,
	"*":  PRODUCT,
	"/":  PRODUCT,
	"%":  PRODUCT,
	"(":  CALL,
	"[":  INDEX,
	"++": POSTFIX, // ⬅️ ADICIONADO
	"--": POSTFIX, // ⬅️ ADICIONADO
}
