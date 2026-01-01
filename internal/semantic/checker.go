package semantic

import (
	"github.com/alpha/internal/parser"
)

type Checker struct {
	CurrentScope *Scope
	Errors       []SemanticError

	// Contexto atual
	currentFuncReturnType parser.Type
	inLoop                bool // NOVO: Para validar break/continue
}

func NewChecker() *Checker {
	global := NewScope(nil)

	// Exemplo: Registrar built-ins primitivos no escopo global para evitar "undeclared identifier 'int'"
	// se 'int' for tratado como identificador em algum lugar.

	return &Checker{
		CurrentScope: global,
		Errors:       make([]SemanticError, 0),
		inLoop:       false,
	}
}

func (c *Checker) CheckProgram(prog *parser.Program) {
	for _, stmt := range prog.Body {
		c.checkStmt(stmt)
	}
}

func (c *Checker) reportError(line, col int, msg string) {
	c.Errors = append(c.Errors, SemanticError{
		Msg:  msg,
		Line: line,
		Col:  col,
	})
}

func (c *Checker) enterScope() {
	c.CurrentScope = NewScope(c.CurrentScope)
}

func (c *Checker) exitScope() {
	if c.CurrentScope.Outer != nil {
		c.CurrentScope = c.CurrentScope.Outer
	}
}
