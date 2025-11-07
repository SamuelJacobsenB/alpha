package parser

import "fmt"

func (p *Parser) parseFunctionDecl() Stmt {
	fmt.Printf("parseFunctionDecl: starting with cur=%q\n", p.cur.Lexeme)

	returnType := p.parseType()
	if returnType == nil {
		p.errorf("expected return type for function")
		return nil
	}

	if !p.expectAndConsume("function") {
		p.errorf("expected 'function' keyword after return type")
		return nil
	}

	name := p.parseFunctionName()
	if name == "" {
		return nil
	}

	generics := p.parseOptionalGenerics()
	params := p.parseFunctionParameters()
	if params == nil {
		return nil
	}

	body := p.parseFunctionBody()
	if body == nil {
		return nil
	}

	return &FunctionDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseFunctionExpr() Expr {
	fmt.Printf("parseFunctionExpr: starting at %q\n", p.cur.Lexeme)

	if !isTypeKeyword(p.cur.Lexeme) {
		p.errorf("expected return type for function expression, got %q", p.cur.Lexeme)
		return nil
	}

	returnType := p.parseType()
	if returnType == nil {
		return nil
	}

	if !p.expectAndConsume("function") {
		p.errorf("expected 'function' keyword")
		return nil
	}

	generics := p.parseOptionalGenerics()
	params := p.parseFunctionParameters()
	if params == nil {
		return nil
	}

	body := p.parseFunctionBody()
	if body == nil {
		return nil
	}

	return &FunctionExpr{
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}
