package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme == "{" {
		return p.parseBlock()
	}

	// Single statement
	if stmt := p.parseTopLevel(); stmt != nil {
		return []Stmt{stmt}
	}

	return nil
}

func (p *Parser) parseBlock() []Stmt {
	if !p.expectAndConsume("{") {
		return nil
	}

	stmts := make([]Stmt, 0, 5)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if stmt := p.parseTopLevel(); stmt != nil {
			stmts = append(stmts, stmt)
		} else {
			p.advanceToken()
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}
	return stmts
}
