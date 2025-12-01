package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseForTraditional() Stmt {
	if !p.expectAndConsume("(") {
		return nil
	}

	// Parse initializer (pode ser nil se começar com ';')
	var init Stmt
	if p.cur.Lexeme != ";" {
		init = p.parseForLoopInitializer()
		if init == nil && p.cur.Lexeme != ";" {
			p.errorf("failed to parse for loop initializer")
			return nil
		}
	}

	// Consumir o ';' após o initializer
	if !p.expectAndConsume(";") {
		return nil
	}

	// Parse condition (pode ser nil se começar com ';')
	var cond Expr
	if p.cur.Lexeme != ";" {
		cond = p.parseExpression(LOWEST)
		if cond == nil {
			p.errorf("failed to parse for loop condition")
			return nil
		}
	}

	if !p.expectAndConsume(";") {
		return nil
	}

	// Parse post statement (pode ser nil se começar com ')')
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
	// Proteger contra tentar parsear ';' como statement
	if p.cur.Lexeme == ";" {
		return nil
	}

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
	// Salva o estado atual do parser
	savedCur := p.cur
	savedNxt := p.nxt

	defer func() {
		// Restaura o estado original
		p.cur = savedCur
		p.nxt = savedNxt
	}()

	// Avança do token 'for' para o próximo
	if p.cur.Lexeme != "(" {
		return false
	}
	p.advanceToken() // consume '('

	// Verifica se temos um identificador
	if p.cur.Type != lexer.IDENT {
		return false
	}
	p.advanceToken() // consume primeiro identificador

	// Verifica se há vírgula (dois identificadores)
	if p.cur.Lexeme == "," {
		p.advanceToken() // consume ','
		if p.cur.Type != lexer.IDENT {
			return false
		}
		p.advanceToken() // consume segundo identificador
	}

	// Agora deve ser 'in'
	return p.cur.Lexeme == "in"
}
