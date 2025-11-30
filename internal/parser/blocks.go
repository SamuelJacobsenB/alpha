package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme != "{" {
		if stmt := p.parseTopLevel(); stmt != nil {
			return []Stmt{stmt}
		}
		return nil
	}
	return p.parseBlock()
}

func (p *Parser) parseBlock() []Stmt {
	p.advanceToken() // consume '{'

	stmts := make([]Stmt, 0, 5) // Pre-aloca

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if stmt := p.parseTopLevel(); stmt != nil {
			stmts = append(stmts, stmt)
		} else {
			p.advanceToken() // Evita loop infinito
		}
	}

	p.expectAndConsume("}")
	return stmts
}
