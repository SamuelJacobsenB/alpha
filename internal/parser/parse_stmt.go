package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTopLevel() Stmt {
	if p.cur.Type == lexer.KEYWORD {
		switch p.cur.Lexeme {
		case "var":
			return p.parseVarDecl()
		case "const":
			return p.parseConstDecl()
		case "if":
			return p.parseIf()
		case "while":
			return p.parseWhile()
		case "for":
			return p.parseFor()
		case "return":
			return p.parseReturn()
		}
	}

	if p.cur.Lexeme == "{" {
		return &BlockStmt{Body: p.parseBlockLike()}
	}

	if p.cur.Lexeme == "}" {
		return nil
	}

	return p.parseExprStmt()
}

func (p *Parser) parseVarDecl() Stmt {
	p.advanceToken() // consome 'var'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after var at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken() // consome identificador

	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken() // consome '='
		init = p.parseExpression(LOWEST)
	}

	return &VarDecl{Name: name, Init: init}
}

func (p *Parser) parseConstDecl() Stmt {
	p.advanceToken() // consome 'const'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after const at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken() // consome identificador

	if p.cur.Lexeme != "=" {
		p.errorf("expected = in const declaration at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	p.advanceToken() // consome '='
	init := p.parseExpression(LOWEST)

	return &ConstDecl{Name: name, Init: init}
}

func (p *Parser) parseExprStmt() Stmt {
	ex := p.parseExpression(LOWEST)
	return &ExprStmt{Expr: ex}
}

func (p *Parser) parseReturn() Stmt {
	p.advanceToken() // consome 'return'

	if p.cur.Type == lexer.EOF || p.cur.Lexeme == "}" {
		return &ReturnStmt{Value: nil}
	}

	val := p.parseExpression(LOWEST)
	return &ReturnStmt{Value: val}
}

func (p *Parser) parseIf() Stmt {
	p.advanceToken() // consome 'if'

	// Consumir '(' se presente
	hasParen := false
	if p.cur.Lexeme == "(" {
		hasParen = true
		p.advanceToken() // consome '('
	}

	cond := p.parseExpression(LOWEST)
	if cond == nil {
		p.errorf("invalid condition in if statement")
		return nil
	}

	// Consumir ')' se tínhamos '('
	if hasParen {
		if p.cur.Lexeme == ")" {
			p.advanceToken() // consome ')'
		} else {
			p.errorf("expected ')' after if condition at %d:%d", p.cur.Line, p.cur.Col)
			return nil
		}
	}

	thenBlock := p.parseBlockLike()

	var elseBlock []Stmt
	if p.cur.Lexeme == "else" {
		p.advanceToken() // consome 'else'
		elseBlock = p.parseBlockLike()
	}

	return &IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}
}

func (p *Parser) parseWhile() Stmt {
	p.advanceToken() // consome 'while'

	if p.cur.Lexeme == "(" {
		p.advanceToken()
	}

	cond := p.parseExpression(LOWEST)

	if p.cur.Lexeme == ")" {
		p.advanceToken() // consome ')'
	}

	body := p.parseBlockLike()
	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseFor() Stmt {
	p.advanceToken() // consome 'for'

	var init Stmt
	var cond Expr
	var post Stmt

	if p.cur.Lexeme == "(" {
		p.advanceToken() // consome '('

		// init
		if p.cur.Lexeme == "var" {
			init = p.parseVarDecl()
		} else if p.cur.Lexeme != ";" {
			init = p.parseExprStmt()
		}

		if p.cur.Lexeme == ";" {
			p.advanceToken() // consome ';'
		}

		// cond
		if p.cur.Lexeme != ";" {
			cond = p.parseExpression(LOWEST)
		}
		if p.cur.Lexeme == ";" {
			p.advanceToken() // consome ';'
		}

		// post
		if p.cur.Lexeme != ")" {
			post = p.parseExprStmt()
		}

		if p.nxt.Lexeme == ")" {
			p.advanceToken()
		}
	}

	if p.cur.Lexeme != "{" && p.nxt.Lexeme == "{" {
		p.advanceToken()
	}

	body := p.parseBlockLike()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseBlockLike() []Stmt {
	if p.cur.Lexeme != "{" {
		stmt := p.parseTopLevel()
		if stmt != nil {
			return []Stmt{stmt}
		}
		return nil
	}

	p.advanceToken() // consome '{'

	var stmts []Stmt
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		stmt := p.parseTopLevel()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}

		if p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
			if !p.isAtStatementBoundary() {
				p.advanceToken()
			}
		}
	}

	if p.cur.Lexeme == "}" {
		p.advanceToken()
	}

	return stmts
}
