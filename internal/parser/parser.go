package parser

import (
	"fmt"
	"strconv"

	"github.com/alpha/internal/lexer"
)

type Parser struct {
	sc     *lexer.Scanner
	cur    lexer.Token
	nxt    lexer.Token
	Errors []string
}

func New(sc *lexer.Scanner) *Parser {
	p := &Parser{sc: sc, Errors: []string{}}
	p.cur = p.sc.NextToken()
	p.nxt = p.sc.NextToken()
	return p
}

func (p *Parser) advanceToken() {
	if p.cur.Type == lexer.EOF {
		return
	}

	p.cur = p.nxt
	p.nxt = p.sc.NextToken()
	fmt.Printf("advanceToken: %q -> %q (nxt=%q)\n", p.cur.Lexeme, p.nxt.Lexeme, p.nxt.Lexeme)
}

func (p *Parser) curType() lexer.TokenType  { return p.cur.Type }
func (p *Parser) peekType() lexer.TokenType { return p.nxt.Type }

func (p *Parser) expectLexeme(lex string) bool {
	if p.nxt.Lexeme == lex {
		p.advanceToken()
		return true
	}
	p.errorf("expected %q, found %q at %d:%d", lex, p.nxt.Lexeme, p.nxt.Line, p.nxt.Col)
	return false
}

func (p *Parser) errorf(format string, a ...interface{}) {
	p.Errors = append(p.Errors, fmt.Sprintf(format, a...))
}

func (p *Parser) ParseProgram() *Program {
	prog := &Program{Body: []Stmt{}}

	for p.curType() != lexer.EOF && p.cur.Lexeme != "" {
		fmt.Printf("=== Iteração %d: cur=%q nxt=%q ===\n", len(prog.Body), p.cur.Lexeme, p.nxt.Lexeme)

		previousToken := p.cur.Lexeme

		stmt := p.parseTopLevel()
		if stmt != nil {
			prog.Body = append(prog.Body, stmt)
			fmt.Printf("=== Adicionado: %T ===\n", stmt)
		}

		if stmt == nil && p.cur.Lexeme == previousToken && p.curType() != lexer.EOF {
			fmt.Printf("=== Token preso em %q, avançando manualmente ===\n", p.cur.Lexeme)
			p.advanceToken()
		}

		if p.curType() == lexer.EOF {
			break
		}
	}

	fmt.Printf("=== Programa completo com %d statements ===\n", len(prog.Body))
	return prog
}

func (p *Parser) isAtStatementBoundary() bool {
	if p.cur.Type == lexer.KEYWORD {
		switch p.cur.Lexeme {
		case "var", "const", "if", "while", "for", "return":
			return true
		}

		if isTypeKeyword(p.cur.Lexeme) {
			return true
		}
	}

	if p.cur.Lexeme == "}" {
		return true
	}

	return false
}

func (p *Parser) parseNumberToken(tok lexer.Token) Expr {
	if tok.Type == lexer.INT {
		v, _ := strconv.ParseInt(tok.Value, 10, 64)
		return &IntLiteral{Value: v}
	}

	f, _ := strconv.ParseFloat(tok.Value, 64)
	return &FloatLiteral{Value: f}
}
