package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseForTraditional() Stmt {
	p.advanceToken()
	if !p.expectAndConsume("(") {
		return nil
	}
	var init Stmt
	if p.cur.Lexeme != ";" {
		init = p.parseForLoopInitializer()
	}
	if !p.expectAndConsume(";") {
		return nil
	}
	var cond Expr
	if p.cur.Lexeme != ";" {
		cond = p.parseExpression(LOWEST)
	}
	if !p.expectAndConsume(";") {
		return nil
	}
	var post Stmt
	if p.cur.Lexeme != ")" {
		postExpr := p.parseExpression(LOWEST)
		if postExpr != nil {
			post = &ExprStmt{Expr: postExpr}
		}
	}
	if !p.expectAndConsume(")") {
		return nil
	}
	body := p.parseBlockLike()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseForIn() Stmt {
	p.advanceToken()
	if !p.expectAndConsume("(") {
		return nil
	}
	index, item := p.parseForInIdentifiers()
	if item == nil {
		return nil
	}
	if !p.expectAndConsume("in") {
		return nil
	}
	iterable := p.parseExpression(LOWEST)
	if iterable == nil {
		return nil
	}
	if !p.expectAndConsume(")") {
		return nil
	}
	body := p.parseBlockLike()
	return &ForInStmt{Index: index, Item: item, Iterable: iterable, Body: body}
}

func (p *Parser) parseForInIdentifiers() (*Identifier, *Identifier) {
	if p.cur.Type != lexer.IDENT {
		return nil, nil
	}
	first := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()
	if p.cur.Lexeme == "," {
		p.advanceToken()
		if p.cur.Type != lexer.IDENT {
			return nil, nil
		}
		second := &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
		return first, second
	}
	return nil, first
}

func (p *Parser) parseForLoopInitializer() Stmt {
	switch {
	case p.cur.Lexeme == "var":
		return p.parseVarDecl()
	case isTypeKeyword(p.cur.Lexeme):
		return p.parseTypedVarDecl()
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) isForInLoop() bool {
	save := p.cur
	defer func() { p.cur = save }()

	p.advanceToken()
	if p.cur.Lexeme != "(" {
		return false
	}
	p.advanceToken()
	count := 0
	for p.cur.Type == lexer.IDENT {
		count++
		p.advanceToken()
		if p.cur.Lexeme == "," {
			p.advanceToken()
			if p.cur.Type != lexer.IDENT {
				return false
			}
			count++
			p.advanceToken()
		}
		break
	}
	if count == 0 {
		return false
	}
	return p.cur.Lexeme == "in"
}
