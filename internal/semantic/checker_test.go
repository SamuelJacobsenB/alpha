package semantic

import (
	"testing"

	"github.com/alpha/internal/lexer"
	"github.com/alpha/internal/parser"
)

// helper para lex+parse+check
func parseAndCheck(t *testing.T, src string) []SemanticErr {
	sc := lexer.NewScanner(src)
	pr := parser.New(sc)
	ast := pr.ParseProgram()
	ch := NewChecker()
	errs := ch.Check(ast)
	return errs
}

func TestUndefinedVariable(t *testing.T) {
	src := `
var x = 10
y = x + 1
`
	errs := parseAndCheck(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected undefined variable error, got none")
	}
	found := false
	for _, e := range errs {
		if e.Msg == "undefined identifier \"y\"" || e.Msg == "assign to undefined variable \"y\"" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected undefined error for y, got: %v", errs)
	}
}

func TestRedeclareLocal(t *testing.T) {
	src := `
var a = 1
var a = 2
`
	errs := parseAndCheck(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected redeclare error, got none")
	}
}

func TestTypeMismatchAssign(t *testing.T) {
	src := `
var x = 1
x = 3.14
`
	errs := parseAndCheck(t, src)
	// assignment int var -> float value should be allowed only if x declared float.
	// Our checker will flag mismatch because x was inferred as int.
	if len(errs) == 0 {
		t.Fatalf("expected type mismatch error, got none")
	}
}

func TestIfConditionType(t *testing.T) {
	src := `
var x = 1
if (x) { x = 2 }
`
	errs := parseAndCheck(t, src)
	found := false
	for _, e := range errs {
		if e.Msg == "if condition must be boolean" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected boolean condition error, got: %v", errs)
	}
}
