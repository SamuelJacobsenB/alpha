package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseFunctionDecl(generic bool) Stmt {
	var generics []*GenericParam
	if generic {
		generics = p.parseGenericParams()
	}

	returnType := p.parseType()
	if !p.expectAndConsume("function") {
		return nil
	}

	name := p.cur.Lexeme
	if p.cur.Type != lexer.IDENT {
		return nil
	}
	p.advanceToken()

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &FunctionDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseFunctionExpr() Expr {
	returnType := p.parseType()
	if !p.expectAndConsume("function") {
		return nil
	}
	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()
	return &FunctionExpr{
		ReturnType: returnType,
		Params:     params,
		Body:       body,
	}
}
