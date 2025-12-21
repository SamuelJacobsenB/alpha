package parser

import (
	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTypedVarDecl() Stmt {
	typ := p.parseType()
	if typ == nil || p.cur.Type != lexer.IDENT {
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		if init = p.parseExpression(LOWEST); init == nil {
			p.errorf("expected expression after '='")
			return nil
		}
	}

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

	var typ Type
	if p.cur.Lexeme == ":" {
		p.advanceToken()
		typ = p.parseType()
		if typ == nil {
			return nil
		}
	}

	return &VarDecl{Name: name, Type: typ, Init: p.parseOptionalInitializer()}
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
	if p.cur.Lexeme != "=" {
		return nil
	}
	p.advanceToken()

	switch p.cur.Lexeme {
	case "{":
		return p.parseMapLiteral()
	case "[":
		return p.parseArrayOrSetLiteral()
	case "&":
		return p.parseReferenceExpr()
	default:
		return p.parseExpression(LOWEST)
	}
}

func (p *Parser) parseArrayOrSetLiteral() Expr {
	p.advanceToken() // consume '['

	elements := p.parseArrayElements()
	if elements == nil || !p.expectAndConsume("]") {
		p.errorf("expected ']' after array/set literal")
		return nil
	}

	return &ArrayLiteral{Elements: elements}
}

func (p *Parser) parseFunctionDecl(generic bool) Stmt {
	var generics []*GenericParam
	if generic {
		generics = p.parseGenericParams()
	}

	returnType := p.parseType()
	if returnType == nil || !p.expectAndConsume("function") || p.cur.Type != lexer.IDENT {
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &FunctionDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseFunctionParameters() []*Param {
	if !p.expectAndConsume("(") {
		return nil
	}

	if p.cur.Lexeme == ")" {
		p.advanceToken()
		return []*Param{}
	}

	params := make([]*Param, 0, 4)

	for {
		paramType := p.parseType()
		if paramType == nil || p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name")
			return nil
		}

		params = append(params, &Param{
			Name: p.cur.Lexeme,
			Type: paramType,
		})
		p.advanceToken()

		if p.cur.Lexeme == ")" {
			p.advanceToken()
			break
		}

		if !p.expectAndConsume(",") {
			return nil
		}
	}

	return params
}

func (p *Parser) parseFunctionBody() []Stmt {
	if p.cur.Lexeme != "{" {
		p.errorf("expected '{' for function body")
		return nil
	}
	return p.parseBlock()
}

// Nova função para parsear declaração genérica: generic<T> T function identity(T value)
func (p *Parser) parseGenericFunctionDecl() Stmt {
	// Já estamos no token "generic"
	p.advanceToken() // consume 'generic'

	// Parsear parâmetros genéricos: <T> ou <T, U>
	if !p.expectAndConsume("<") {
		return nil
	}

	generics := p.parseGenericList()
	if generics == nil || !p.expectAndConsume(">") {
		return nil
	}

	// Parsear tipo de retorno (pode ser um tipo genérico como T)
	returnType := p.parseType()
	if returnType == nil || !p.expectAndConsume("function") || p.cur.Type != lexer.IDENT {
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &FunctionDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseGenericParams() []*GenericParam {
	if !p.expectAndConsume("<") {
		return nil
	}

	generics := p.parseGenericList()
	if generics == nil || !p.expectAndConsume(">") {
		return nil
	}

	return generics
}

func (p *Parser) parseGenericList() []*GenericParam {
	generics := make([]*GenericParam, 0, 2)

	// Parse first parameter - aceita tanto IDENT quanto GENERIC (T, U)
	if p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
		p.errorf("expected identifier or generic parameter, got %s", p.cur.Lexeme)
		return nil
	}

	generics = append(generics, &GenericParam{Name: p.cur.Lexeme})
	p.advanceToken()

	// Parse additional parameters
	for p.cur.Lexeme == "," {
		p.advanceToken()

		if p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
			p.errorf("expected identifier or generic parameter, got %s", p.cur.Lexeme)
			return nil
		}

		generics = append(generics, &GenericParam{Name: p.cur.Lexeme})
		p.advanceToken()
	}

	return generics
}

func (p *Parser) parseClass() Stmt {
	p.advanceToken() // consume 'class'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected class name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	if !p.expectAndConsume("{") {
		return nil
	}

	fields := make([]*FieldDecl, 0, 4)
	var constructor *ConstructorDecl
	methods := make([]*MethodDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		switch p.cur.Lexeme {
		case "constructor":
			constructor = p.parseConstructor()
		default:
			if p.isMethodDeclaration() {
				method := p.parseMethod()
				if method != nil {
					methods = append(methods, method)
				}
			} else {
				field := p.parseField()
				if field != nil {
					fields = append(fields, field)
				}
			}
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &ClassDecl{
		Name:        name,
		Generics:    generics,
		Fields:      fields,
		Constructor: constructor,
		Methods:     methods,
	}
}

func (p *Parser) parseField() *FieldDecl {
	typ := p.parseType()
	if typ == nil || p.cur.Type != lexer.IDENT {
		p.errorf("expected field name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()
	p.consumeOptionalSemicolon()

	return &FieldDecl{Name: name, Type: typ}
}

func (p *Parser) parseConstructor() *ConstructorDecl {
	p.advanceToken() // consume 'constructor'

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &ConstructorDecl{Params: params, Body: body}
}

func (p *Parser) parseMethod() *MethodDecl {
	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	returnType := p.parseType()
	if returnType == nil || !p.expectAndConsume("method") || p.cur.Type != lexer.IDENT {
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &MethodDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) isMethodDeclaration() bool {
	saved := p.cur
	defer func() { p.cur = saved }()

	// Check generics
	if p.cur.Lexeme == "<" {
		p.advanceToken()

		if p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
			return false
		}
		p.advanceToken()

		for p.cur.Lexeme == "," {
			p.advanceToken()
			if p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
				return false
			}
			p.advanceToken()
		}

		if p.cur.Lexeme != ">" {
			return false
		}
		p.advanceToken()
	}

	// Check return type
	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT && p.cur.Type != lexer.GENERIC {
		return false
	}
	p.advanceToken()

	// Check type modifiers
	for p.cur.Lexeme == "?" || p.cur.Lexeme == "*" || p.cur.Lexeme == "[" {
		if p.cur.Lexeme == "[" {
			p.advanceToken()
			if p.cur.Lexeme == "]" {
				p.advanceToken()
			}
		} else {
			p.advanceToken()
		}
	}

	return p.cur.Lexeme == "method"
}

func (p *Parser) parseTypeDecl() Stmt {
	p.advanceToken() // consume 'type'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name after 'type'")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	typ := p.parseTypeBody()
	if typ == nil {
		return nil
	}

	return &TypeDecl{Name: name, Generics: generics, Type: typ}
}

func (p *Parser) parseTypeBody() Type {
	if p.cur.Lexeme == "{" {
		return p.parseStructType()
	}
	return p.parseType()
}

func (p *Parser) parseStructType() Type {
	p.advanceToken() // consume '{'

	fields := make([]*FieldDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		field := p.parseField()
		if field != nil {
			fields = append(fields, field)
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructType{Fields: fields}
}
