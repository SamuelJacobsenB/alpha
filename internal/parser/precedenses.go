package parser

const (
	_ int = iota
	LOWEST
	TERNARY    // ? :
	ASSIGNMENT // = += -= etc.
	LOGICALOR  // ||
	LOGICALAND // &&
	EQUALITY   // == !=
	COMPARISON // < > <= >=
	SUM        // + -
	PRODUCT    // * / %
	PREFIX     // -X !X ++X --X &X
	CALL       // func(...)
	MEMBER     // obj.member
	INDEX      // [] - acesso a array
	POSTFIX    // X++ X--
)

var precedences = map[string]int{
	"?":  TERNARY,
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
	".":  MEMBER,
	"[":  INDEX,
	"++": POSTFIX,
	"--": POSTFIX,
}
