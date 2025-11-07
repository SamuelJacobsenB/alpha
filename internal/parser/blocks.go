package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme != "{" {
		return p.parseSingleStatement()
	}

	return p.parseBlock()
}

func (p *Parser) parseSingleStatement() []Stmt {
	stmt := p.parseTopLevel()
	if stmt != nil {
		return []Stmt{stmt}
	}
	return nil
}

func (p *Parser) parseBlock() []Stmt {
	p.advanceToken() // consume '{'
	fmt.Printf("parseBlock: entered block, cur=%q\n", p.cur.Lexeme)

	var stmts []Stmt

	for !p.isBlockEnd() && p.cur.Type != lexer.EOF {
		stmt := p.parseStatementInBlock()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	if !p.expectAndConsume("}") {
		p.errorf("expected '}' to close block, got %q", p.cur.Lexeme)
	}

	fmt.Printf("parseBlock: exited block with %d statements\n", len(stmts))
	return stmts
}

func (p *Parser) parseStatementInBlock() Stmt {
	previousToken := p.cur.Lexeme
	stmt := p.parseTopLevel()

	// Safety: advance if we're stuck on the same token
	if stmt == nil && p.cur.Lexeme == previousToken && !p.isBlockEnd() {
		p.advanceToken()
	}

	return stmt
}

func (p *Parser) parseExprStmt() Stmt {
	ex := p.parseExpression(LOWEST)
	return &ExprStmt{Expr: ex}
}
