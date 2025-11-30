package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseGenericParams() []*GenericParam {
	if !p.expectAndConsume("<") {
		return nil
	}
	var generics []*GenericParam
	for p.cur.Type == lexer.GENERIC {
		generics = append(generics, &GenericParam{Name: p.cur.Lexeme})
		p.advanceToken()
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != ">" {
			p.errorf("expected ',' or '>'")
			return nil
		}
	}
	if !p.expectAndConsume(">") {
		return nil
	}
	return generics
}

func (p *Parser) isGenericCall() bool {
	save := p.cur
	defer func() { p.cur = save }()

	if p.cur.Lexeme != "<" {
		return false
	}
	p.advanceToken()

	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT {
		return false
	}

	for p.cur.Lexeme != ">" && p.cur.Type != lexer.EOF {
		if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT && p.cur.Lexeme != "," {
			return false
		}
		p.advanceToken()
	}

	if p.cur.Lexeme != ">" {
		return false
	}
	p.advanceToken()

	return p.cur.Lexeme == "("
}

func (p *Parser) parseGenericCall(left Expr) Expr {
	p.advanceToken()
	typeArgs := p.parseTypeArguments()
	if typeArgs == nil {
		return nil
	}
	if !p.expectAndConsume(">") {
		return nil
	}
	if p.cur.Lexeme != "(" {
		p.errorf("expected '(' after generic arguments")
		return nil
	}
	call := p.parseCall(left)
	if call == nil {
		return nil
	}
	return &GenericSpecialization{
		Callee:   left,
		TypeArgs: typeArgs,
	}
}

func (p *Parser) parseTypeArguments() []Type {
	var typeArgs []Type
	for {
		typ := p.parseType()
		if typ == nil {
			return nil
		}
		typeArgs = append(typeArgs, typ)
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else {
			break
		}
	}
	return typeArgs
}
