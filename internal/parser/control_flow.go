package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseKeywordStatement() Stmt {
	switch p.cur.Lexeme {
	case "var":
		fmt.Println("parseTopLevel: detected VAR")
		return p.parseVarDecl()
	case "const":
		fmt.Println("parseTopLevel: detected CONST")
		return p.parseConstDecl()
	case "function":
		return p.handleTopLevelFunction()
	case "if":
		fmt.Println("parseTopLevel: detected IF")
		return p.parseIf()
	case "while":
		fmt.Println("parseTopLevel: detected WHILE")
		return p.parseWhile()
	case "for":
		fmt.Println("parseTopLevel: detected FOR")
		return p.parseFor()
	case "return":
		fmt.Println("parseTopLevel: detected RETURN")
		return p.parseReturn()
	default:
		p.advanceToken()
		return nil
	}
}

func (p *Parser) handleTopLevelFunction() Stmt {
	p.errorf("anonymous functions are not allowed at top level")
	p.advanceToken() // skip 'function'
	return nil
}

func (p *Parser) parseIf() Stmt {
	p.advanceToken() // consume 'if'

	cond := p.parseCondition()
	if cond == nil {
		return nil
	}

	thenBlock := p.parseBlockLike()
	elseBlock := p.parseOptionalElse()

	return &IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}
}

func (p *Parser) parseWhile() Stmt {
	p.advanceToken() // consume 'while'

	cond := p.parseCondition()
	body := p.parseBlockLike()

	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseFor() Stmt {
	fmt.Printf("parseFor: starting, cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)

	if p.isForInLoop() {
		fmt.Println("parseFor: detected for...in loop")
		return p.parseForIn()
	}

	fmt.Println("parseFor: detected traditional for loop")
	return p.parseForTraditional()
}

func (p *Parser) parseReturn() Stmt {
	p.advanceToken() // consume 'return'

	if p.isAtEndOfStatement() {
		return &ReturnStmt{Value: nil}
	}

	value := p.parseExpression(LOWEST)
	return &ReturnStmt{Value: value}
}

// Helper functions
func (p *Parser) parseCondition() Expr {
	hasParen := p.cur.Lexeme == "("
	if hasParen {
		p.advanceToken() // consume '('
	}

	cond := p.parseExpression(LOWEST)
	if cond == nil {
		p.errorf("invalid condition")
		return nil
	}

	if hasParen {
		if !p.expectAndConsume(")") {
			p.errorf("expected ')' after condition")
			return nil
		}
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

func (p *Parser) isForInLoop() bool {
	if p.nxt.Lexeme != "(" {
		return false
	}

	saveCur := p.cur
	saveNxt := p.nxt

	p.advanceToken() // cur = "("
	p.advanceToken() // cur = primeiro token dentro dos parênteses

	isForIn := false
	steps := 0

	for steps < 5 && p.cur.Type != lexer.EOF && p.cur.Lexeme != ")" {
		if p.cur.Lexeme == "in" {
			isForIn = true
			break
		}
		p.advanceToken()
		steps++
	}

	p.cur = saveCur
	p.nxt = saveNxt

	return isForIn
}

func (p *Parser) isAtEndOfStatement() bool {
	return p.cur.Type == lexer.EOF || p.cur.Lexeme == "}" || p.cur.Lexeme == ";"
}

func (p *Parser) expectAndConsume(expected string) bool {
	if p.cur.Lexeme == expected {
		p.advanceToken()
		return true
	}
	return false
}
