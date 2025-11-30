package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseControlStmt() Stmt {
	switch p.cur.Lexeme {
	case "if":
		return p.parseIf()
	case "while":
		return p.parseWhile()
	case "for":
		return p.parseFor()
	case "return":
		return p.parseReturn()
	default:
		p.advanceToken()
		return nil
	}
}

func (p *Parser) parseIf() Stmt {
	p.advanceToken()

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	thenBlock := p.parseBlockLike()
	elseBlock := p.parseOptionalElse()

	return &IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}
}

func (p *Parser) parseWhile() Stmt {
	p.advanceToken()

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	return &WhileStmt{Cond: cond, Body: p.parseBlockLike()}
}

func (p *Parser) parseFor() Stmt {
	p.advanceToken()

	if p.isForInLoop() {
		return p.parseForIn()
	}
	return p.parseForTraditional()
}

func (p *Parser) parseReturn() Stmt {
	p.advanceToken()

	if p.isAtEndOfStatement() {
		return &ReturnStmt{Value: nil}
	}

	return &ReturnStmt{Value: p.parseExpression(LOWEST)}
}

func (p *Parser) parseCondition() Expr {
	if !p.expectAndConsume("(") {
		return nil
	}

	cond := p.parseExpression(LOWEST)
	if !p.expectAndConsume(")") {
		return nil
	}

	return cond
}

func (p *Parser) parseOptionalElse() []Stmt {
	if p.cur.Lexeme == "else" {
		p.advanceToken()
		return p.parseBlockLike()
	}
	return nil
}

func (p *Parser) isAtEndOfStatement() bool {
	return p.cur.Lexeme == ";" || p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF
}

func (p *Parser) expectAndConsume(expected string) bool {
	if p.cur.Lexeme == expected {
		p.advanceToken()
		return true
	}
	p.errorf("expected '%s', got '%s'", expected, p.cur.Lexeme)
	return false
}
