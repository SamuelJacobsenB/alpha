package parser

import "github.com/alpha/internal/lexer"

func (p *Parser) parseControlStmt() Stmt {
	switch p.cur.Lexeme {
	case "if":
		return p.parseIf()
	case "while":
		return p.parseWhile()
	case "do":
		return p.parseDoWhile()
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

func (p *Parser) parseDoWhile() Stmt {
	p.advanceToken() // consume 'do'

	body := p.parseBlockLike()
	if body == nil {
		return nil
	}

	if !p.expectAndConsume("while") {
		p.errorf("expected 'while' after do block")
		return nil
	}

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	return &DoWhileStmt{Body: body, Cond: cond}
}

func (p *Parser) parseFor() Stmt {
	p.advanceToken() // consume 'for'

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
		p.errorf("expected '(' after %s", p.cur.Lexeme)
		return nil
	}

	cond := p.parseExpression(LOWEST)
	if cond == nil {
		p.errorf("expected condition expression")
		p.syncToNextStmt()
		return nil
	}

	if !p.expectAndConsume(")") {
		p.errorf("expected ')' after condition")
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
