package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseForTraditional() Stmt {
	fmt.Printf("parseForTraditional: starting\n")
	p.advanceToken() // consume 'for'

	var init Stmt
	var cond Expr
	var post Stmt

	if p.cur.Lexeme == "(" {
		p.advanceToken() // consume '('
		init, cond, post = p.parseForLoopParts()

		if !p.expectAndConsume(")") {
			p.errorf("expected ')' after for loop")
			return nil
		}
	}

	body := p.parseBlockLike()

	fmt.Printf("parseForTraditional: completed\n")
	return &ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}
}

func (p *Parser) parseForLoopParts() (init Stmt, cond Expr, post Stmt) {
	// Parse initializer
	if p.cur.Lexeme != ";" {
		init = p.parseForLoopInitializer()
	}

	if !p.expectAndConsume(";") {
		p.errorf("expected ';' after for loop initializer")
		return
	}

	// Parse condition
	if p.cur.Lexeme != ";" {
		cond = p.parseExpression(LOWEST)
	}

	if !p.expectAndConsume(";") {
		p.errorf("expected ';' after for loop condition")
		return
	}

	// Parse post statement
	if p.cur.Lexeme != ")" {
		postExpr := p.parseExpression(LOWEST)
		if postExpr != nil {
			post = &ExprStmt{Expr: postExpr}
		}
	}

	return
}

func (p *Parser) parseForLoopInitializer() Stmt {
	switch {
	case p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme):
		return p.parseTypedVarDecl()
	case p.cur.Lexeme == "var":
		return p.parseVarDecl()
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) parseForIn() Stmt {
	fmt.Printf("parseForIn: starting\n")
	p.advanceToken() // consume 'for'

	if !p.expectAndConsume("(") {
		p.errorf("expected '(' after 'for' in for...in loop")
		return nil
	}

	indexIdent, itemIdent := p.parseForInIdentifiers()
	if itemIdent == nil {
		return nil
	}

	if !p.expectAndConsume("in") {
		p.errorf("expected 'in' in for...in loop")
		return nil
	}

	iterable := p.parseExpression(LOWEST)
	if iterable == nil {
		p.errorf("invalid iterable expression in for...in loop")
		return nil
	}

	if !p.expectAndConsume(")") {
		p.errorf("expected ')' after for...in loop")
		return nil
	}

	body := p.parseBlockLike()

	fmt.Printf("parseForIn: completed %s in %T\n", itemIdent.Name, iterable)
	return &ForInStmt{
		Index:    indexIdent,
		Item:     itemIdent,
		Iterable: iterable,
		Body:     body,
	}
}

func (p *Parser) parseForInIdentifiers() (*Identifier, *Identifier) {
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier in for...in loop")
		return nil, nil
	}

	firstIdent := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()

	if p.cur.Lexeme == "," {
		p.advanceToken() // consume ','

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected second identifier after ',' in for...in loop")
			return nil, nil
		}

		return firstIdent, &Identifier{Name: p.cur.Lexeme}
	}

	return nil, firstIdent
}
