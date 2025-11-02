package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTopLevel() Stmt {
	fmt.Printf("parseTopLevel enter cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)

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
	p.advanceToken() // consome 'return' => p.cur agora é início da expressão de retorno (ou EOF)
	if p.cur.Type == lexer.EOF {
		return &ReturnStmt{Value: nil}
	}
	val := p.parseExpression(LOWEST)
	// NÃO avance aqui: deixe p.cur apontar para o último token do retorno
	return &ReturnStmt{Value: val}
}

func (p *Parser) parseIf() Stmt {
	// entrada: p.cur == "if"
	p.advanceToken() // consome 'if', agora p.cur é token após 'if'

	// consumir '(' se estiver presente e posicionar p.cur no primeiro token da condição
	if p.cur.Lexeme == "(" {
		p.advanceToken() // consome '('
	}

	// parse da condição (deve deixar p.cur no último token da condição)
	cond := p.parseExpression(LOWEST)

	// consumir ')' se ele estiver em p.cur ou em p.nxt
	if p.cur.Lexeme == ")" {
		p.advanceToken() // consome ')', p.cur -> token após ')'
	} else if p.nxt.Lexeme == ")" {
		p.advanceToken() // posiciona p.cur == ')'
		p.advanceToken() // consome ')', p.cur -> token após ')'
	}

	// se o bloco começa com '{' e ele está em p.nxt, posicione p.cur nele
	if p.cur.Lexeme != "{" && p.nxt.Lexeme == "{" {
		p.advanceToken() // posiciona p.cur == '{'
	}

	then := p.parseBlockLike()

	var els []Stmt
	// detectar else (pode estar em p.cur ou em p.nxt)
	if p.cur.Lexeme == "else" || p.nxt.Lexeme == "else" {
		// se else está em p.nxt, avance para ele
		if p.nxt.Lexeme == "else" {
			p.advanceToken() // p.cur == 'else'
		}
		// agora p.cur == 'else', consumir e posicionar p.cur no token após 'else'
		p.advanceToken() // consome 'else', p.cur -> token após 'else'

		// se else começar com bloco e '{' estiver em p.nxt, posicione p.cur no '{'
		if p.cur.Lexeme != "{" && p.nxt.Lexeme == "{" {
			p.advanceToken() // posiciona p.cur == '{'
		}

		// parseBlockLike deixará p.cur == '}' ao retornar (se bloco) ou p.cur no último token do stmt simples
		els = p.parseBlockLike()
	}

	return &IfStmt{Cond: cond, Then: then, Else: els}
}

func (p *Parser) parseWhile() Stmt {
	p.advanceToken() // consome 'while'

	if p.cur.Lexeme == "(" {
		p.advanceToken()
	}
	cond := p.parseExpression(LOWEST)

	// consumir ')'
	if p.cur.Lexeme == ")" {
		p.advanceToken()
	} else if p.nxt.Lexeme == ")" {
		p.advanceToken()
		p.advanceToken()
	}

	// posicionar para bloco se necessário
	if p.cur.Lexeme != "{" && p.nxt.Lexeme == "{" {
		p.advanceToken()
	}
	body := p.parseBlockLike()
	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseFor() Stmt {
	p.advanceToken() // consome 'for'

	var init Stmt
	var cond Expr
	var post Stmt

	// suporte a for ( ... )
	if p.cur.Lexeme == "(" {
		p.advanceToken() // consome '(' -> p.cur é o primeiro token dentro dos parênteses

		// init
		if p.cur.Lexeme == "var" {
			init = p.parseVarDecl()
		} else if p.cur.Lexeme != ";" {
			init = p.parseExprStmt()
		}
		// se o init retornou nil, parseVarDecl/parseExprStmt podem ter deixado p.cur apropriado
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
		// tentar consumir ')' quer esteja em p.cur ou em p.nxt
		if p.cur.Lexeme == ")" {
			// cur já é ')', deixamos como está (o chamador decide o avanço)
		} else if p.nxt.Lexeme == ")" {
			p.advanceToken() // posiciona p.cur == ')'
		}
	}

	// posicionar p.cur no '{' se o bloco estiver em p.nxt
	if p.cur.Lexeme != "{" && p.nxt.Lexeme == "{" {
		p.advanceToken()
	}

	body := p.parseBlockLike()
	return &ForStmt{Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseBlockLike() []Stmt {
	// bloco { ... } : função consome '{' e PARSE os statements internos,
	// mas NÃO consome o '}' ao retornar — deixa p.cur == '}'.
	if p.cur.Lexeme == "{" {
		p.advanceToken() // consome '{', p.cur == primeiro token dentro do bloco (ou '}')

		var stmts []Stmt
		for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
			stmt := p.parseTopLevel()
			if stmt != nil {
				stmts = append(stmts, stmt)
			}

			// Se parseTopLevel deixou p.cur no '}' ou EOF, saia
			if p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF {
				break
			}

			// Caso contrário, avance para o próximo statement interno
			p.advanceToken()
		}

		// NÃO consome '}' aqui. Deixa p.cur == '}' ao retornar.
		return stmts
	}

	// single statement: parseTopLevel espera p.cur no início do statement
	stmt := p.parseTopLevel()
	return []Stmt{stmt}
}
