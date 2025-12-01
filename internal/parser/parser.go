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
	p.advanceToken()
	p.advanceToken()
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
		if stmt := p.parseTopLevel(); stmt != nil {
			body = append(body, stmt)
		} else {
			p.advanceToken()
		}
	}

	return &Program{Body: body}
}

func (p *Parser) parseNumberToken(tok lexer.Token) Expr {
	if tok.Type == lexer.INT {
		v, _ := strconv.ParseInt(tok.Value, 10, 64)
		p.advanceToken()
		return &IntLiteral{Value: v}
	}
	f, _ := strconv.ParseFloat(tok.Value, 64)
	p.advanceToken()
	return &FloatLiteral{Value: f}
}
