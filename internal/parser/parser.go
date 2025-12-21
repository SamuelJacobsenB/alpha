package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alpha/internal/lexer"
)

type Parser struct {
	sc     *lexer.Scanner
	cur    lexer.Token
	nxt    lexer.Token
	Errors []string
}

func New(sc *lexer.Scanner) *Parser {
	p := &Parser{sc: sc}
	p.advanceToken() // Carrega primeiro token em cur
	p.advanceToken() // Carrega segundo token em nxt
	return p
}

func (p *Parser) advanceToken() {
	p.cur = p.nxt
	p.nxt = p.sc.NextToken()
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.Errors = append(p.Errors, fmt.Sprintf(format, args...))
}

func (p *Parser) consumeOptionalSemicolon() {
	if p.cur.Lexeme == ";" {
		p.advanceToken()
	}
}

func (p *Parser) ParseProgram() *Program {
	body := make([]Stmt, 0, 10)

	for p.cur.Type != lexer.EOF {
		stmt := p.parseTopLevel()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			// Evita loop infinito quando não consegue parsear
			if p.cur.Type == lexer.EOF {
				break
			}
			// Tenta sincronizar avançando um token
			p.advanceToken()
		}
	}

	return &Program{Body: body}
}

func (p *Parser) parseNumberToken(tok lexer.Token) Expr {
	var expr Expr

	if tok.Type == lexer.INT {
		val, _ := strconv.ParseInt(tok.Value, 10, 64)
		expr = &IntLiteral{Value: val}
	} else {
		val, _ := strconv.ParseFloat(tok.Value, 64)
		expr = &FloatLiteral{Value: val}
	}

	p.advanceToken()
	return expr
}

func (p *Parser) expectAndConsume(expected string) bool {
	if p.cur.Lexeme == expected {
		p.advanceToken()
		return true
	}
	p.errorf("expected '%s', got '%s'", expected, p.cur.Lexeme)
	return false
}

func (p *Parser) isAtEndOfStatement() bool {
	return p.cur.Lexeme == ";" || p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF
}

func (p *Parser) syncToNextStmt() {
	for !p.isAtStmtStart() && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}
}

func (p *Parser) isAtStmtStart() bool {
	return p.cur.Lexeme == ";" ||
		p.cur.Lexeme == "}" ||
		p.cur.Type == lexer.KEYWORD ||
		p.cur.Lexeme == "{" ||
		isTypeKeyword(p.cur.Lexeme)
}

func (p *Parser) HasErrors() bool {
	return len(p.Errors) > 0
}

func (p *Parser) ErrorsText() string {
	return strings.Join(p.Errors, "\n")
}

func (p *Parser) syncTo(token string) {
	for p.cur.Lexeme != token && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}
	if p.cur.Lexeme == token {
		p.advanceToken()
	}
}
