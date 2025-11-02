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
	// inicializa cur e nxt chamando NextToken duas vezes
	p.cur = p.sc.NextToken()
	p.nxt = p.sc.NextToken()
	return p
}

func (p *Parser) advanceToken() {
	// se já estamos no EOF, não avançamos mais
	if p.cur.Type == lexer.EOF {
		return
	}

	// se o próximo já é EOF, apenas posicione cur = nxt e zere nxt sem chamar NextToken
	if p.nxt.Type == lexer.EOF {
		p.cur = p.nxt
		p.nxt = lexer.Token{}
		fmt.Printf("advanceToken -> cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)
		return
	}

	// caso normal: desloca e busca o próximo token
	p.cur = p.nxt
	p.nxt = p.sc.NextToken()
	fmt.Printf("advanceToken -> cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)
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

// ParseProgram consome tokens até EOF e retorna AST
func (p *Parser) ParseProgram() *Program {
	prog := &Program{Body: []Stmt{}}

	for i := 0; p.curType() != lexer.EOF && p.curType() != lexer.ERROR; i++ {
		fmt.Printf("iter %d cur=%q nxt=%q\n", i, p.cur.Lexeme, p.nxt.Lexeme)

		stmt := p.parseTopLevel()
		if stmt != nil {
			prog.Body = append(prog.Body, stmt)
		}

		// se parseTopLevel deixou p.cur em EOF, paramos
		if p.curType() == lexer.EOF || p.cur.Lexeme == "" {
			break
		}

		// se p.nxt é EOF, faça um único advance para posicionar cur==EOF e saia
		if p.nxt.Type == lexer.EOF {
			p.advanceToken()
			break
		}

		fmt.Printf("before final advance cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)
		p.advanceToken()
		fmt.Printf("after final advance cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)
	}
	return prog
}

// utilitário para criar literal numérico a partir do token
func (p *Parser) parseNumberToken(tok lexer.Token) Expr {
	if tok.Type == lexer.INT {
		v, _ := strconv.ParseInt(tok.Value, 10, 64)
		return &IntLiteral{Value: v}
	}
	f, _ := strconv.ParseFloat(tok.Value, 64)
	return &FloatLiteral{Value: f}
}
