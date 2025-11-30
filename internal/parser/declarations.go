package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTypedVarDecl() Stmt {
	fmt.Printf("parseTypedVarDecl: starting with %q\n", p.cur.Lexeme)

	typ := p.parseType()
	if typ == nil {
		return nil
	}

	fmt.Printf("parseTypedVarDecl: parsed type %T, cur=%q\n", typ, p.cur.Lexeme)

	// Verificar se o próximo token é um identificador
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after type at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Só tentar parsear inicializador se houver '='
	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		fmt.Printf("parseOptionalInitializer: after '=', cur=%q (type: %v)\n",
			p.cur.Lexeme, p.cur.Type)

		// Verificar se é um literal de mapa
		switch p.cur.Lexeme {
		case "{":
			fmt.Println("parseOptionalInitializer: detected map literal")
			init = p.parseMapLiteral()
		case "[":
			fmt.Println("parseOptionalInitializer: detected array/set literal")
			init = p.parseArrayOrSetLiteral()
		case "&":
			fmt.Println("parseOptionalInitializer: detected reference expression")
			init = p.parseReferenceExpr()
		default:
			init = p.parseExpression(LOWEST)
			fmt.Printf("parseOptionalInitializer: parsed expression %T\n", init)
		}
	}

	fmt.Printf("parseTypedVarDecl: completed %s %T, cur=%q\n", name, typ, p.cur.Lexeme)
	return &VarDecl{Name: name, Type: typ, Init: init}
}

func (p *Parser) parseVarDecl() Stmt {
	p.advanceToken() // consume 'var'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after var at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Verificar se há tipo explícito (ex: var x: int = 10)
	var typ Type
	if p.cur.Lexeme == ":" {
		p.advanceToken()
		typ = p.parseType()
		if typ == nil {
			return nil
		}
	}

	init := p.parseOptionalInitializer()
	return &VarDecl{Name: name, Type: typ, Init: init}
}

func (p *Parser) parseConstDecl() Stmt {
	p.advanceToken() // consume 'const'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after const at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	if p.cur.Lexeme != "=" {
		p.errorf("expected = in const declaration at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	p.advanceToken()
	init := p.parseExpression(LOWEST)

	return &ConstDecl{Name: name, Init: init}
}

func (p *Parser) parseOptionalInitializer() Expr {
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		fmt.Printf("parseOptionalInitializer: after '=', cur=%q (type: %v)\n",
			p.cur.Lexeme, p.cur.Type)

		// Verificar se é um literal de mapa
		if p.cur.Lexeme == "{" {
			fmt.Println("parseOptionalInitializer: detected map literal")
			return p.parseMapLiteral()
		}

		// Verificar se é um literal de array/set
		if p.cur.Lexeme == "[" {
			fmt.Println("parseOptionalInitializer: detected array/set literal")
			return p.parseArrayOrSetLiteral()
		}

		// Verificar se é uma expressão de referência
		if p.cur.Lexeme == "&" {
			fmt.Println("parseOptionalInitializer: detected reference expression")
			return p.parseReferenceExpr()
		}

		expr := p.parseExpression(LOWEST)
		fmt.Printf("parseOptionalInitializer: parsed expression %T\n", expr)
		return expr
	}
	return nil
}

func (p *Parser) parseArrayOrSetLiteral() Expr {
	fmt.Printf("parseArrayOrSetLiteral: starting at %q\n", p.cur.Lexeme)
	p.advanceToken() // consume '['

	elements := p.parseArrayElements()
	if elements == nil {
		return nil
	}

	if !p.expectAndConsume("]") {
		p.errorf("expected ']' after array/set literal")
		return nil
	}

	fmt.Printf("parseArrayOrSetLiteral: completed with %d elements\n", len(elements))
	return &ArrayLiteral{Elements: elements}
}
