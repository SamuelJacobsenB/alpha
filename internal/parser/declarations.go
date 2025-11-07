package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTypedVarDecl() Stmt {
	fmt.Printf("parseTypedVarDecl: starting\n")

	typ := p.parseType()
	if typ == nil {
		return nil
	}

	fmt.Printf("parseTypedVarDecl: parsed type %T, cur=%q\n", typ, p.cur.Lexeme)

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after type at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	init := p.parseOptionalInitializer()

	fmt.Printf("parseTypedVarDecl: completed %s, cur=%q\n", name, p.cur.Lexeme)
	return &VarDecl{Name: name, Type: typ, Init: init}
}

func (p *Parser) parseVarDecl() Stmt {
	p.advanceToken() // consume 'var'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after var at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	init := p.parseOptionalInitializer()
	return &VarDecl{Name: name, Init: init}
}

func (p *Parser) parseConstDecl() Stmt {
	p.advanceToken() // consume 'const'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after const at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	if p.cur.Lexeme != "=" {
		p.errorf("expected = in const declaration at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	p.advanceToken()
	init := p.parseExpression(LOWEST)

	return &ConstDecl{Name: name, Init: init}
}

func (p *Parser) parseOptionalInitializer() Expr {
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		fmt.Printf("parseOptionalInitializer: after '=', cur=%q\n", p.cur.Lexeme)
		return p.parseExpression(LOWEST)
	}
	return nil
}
