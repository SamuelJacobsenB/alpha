package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTypedVarDecl() Stmt {
	typ := p.parseType()
	if typ == nil {
		return nil
	}

	fmt.Println(typ, "type parsed", p.cur.Lexeme)

	if p.cur.Type != lexer.IDENT {
		if p.cur.Lexeme == ";" {
			return nil
		}

		p.errorf("expected identifier after type, got '%s' at %d:%d", p.cur.Lexeme, p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	fmt.Println("Hello name", p.cur.Lexeme)
	p.advanceToken()

	// Apenas verificamos a inicialização
	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		init = p.parseExpression(LOWEST)
		if init == nil {
			p.errorf("expected expression after '='")
			return nil
		}
	}

	fmt.Println("Hello", init)

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

func (p *Parser) parseFunctionDecl(generic bool) Stmt {
	var generics []*GenericParam
	if generic {
		generics = p.parseGenericParams()
	}

	returnType := p.parseType()
	if !p.expectAndConsume("function") {
		return nil
	}

	name := p.cur.Lexeme
	if p.cur.Type != lexer.IDENT {
		return nil
	}
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

	var params []*Param

	for {
		paramType := p.parseType()
		if paramType == nil {
			return nil
		}

		if p.cur.Type != lexer.IDENT {
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

func (p *Parser) parseGenericParams() []*GenericParam {
	if !p.expectAndConsume("<") {
		return nil
	}
	var generics []*GenericParam
	for p.cur.Type == lexer.GENERIC {
		generics = append(generics, &GenericParam{Name: p.cur.Lexeme})
		p.advanceToken()
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != ">" {
			p.errorf("expected ',' or '>'")
			return nil
		}
	}
	if !p.expectAndConsume(">") {
		return nil
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

	// Generics opcionais: class List<T>
	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	if !p.expectAndConsume("{") {
		return nil
	}

	var fields []*FieldDecl
	var constructor *ConstructorDecl
	var methods []*MethodDecl

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		switch p.cur.Lexeme {
		case "constructor":
			constructor = p.parseConstructor()
		default:
			// Pode ser field ou method
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
	if typ == nil {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected field name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Consumir semicolon opcional
	p.consumeOptionalSemicolon()

	return &FieldDecl{Name: name, Type: typ}
}

func (p *Parser) parseConstructor() *ConstructorDecl {
	p.advanceToken() // consume 'constructor'

	params := p.parseFunctionParameters()
	if params == nil {
		return nil
	}

	body := p.parseFunctionBody()
	if body == nil {
		return nil
	}

	return &ConstructorDecl{Params: params, Body: body}
}

func (p *Parser) parseMethod() *MethodDecl {
	// Pode ter generics: <T> string method name(...)
	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	// Return type
	returnType := p.parseType()
	if returnType == nil {
		return nil
	}

	// Palavra "method"
	if !p.expectAndConsume("method") {
		return nil
	}

	// Nome do método
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected method name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Parâmetros
	params := p.parseFunctionParameters()
	if params == nil {
		return nil
	}

	// Corpo
	body := p.parseFunctionBody()
	if body == nil {
		return nil
	}

	return &MethodDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) isMethodDeclaration() bool {
	// Verifica se é: [<T>] type method identifier
	saved := p.cur
	defer func() { p.cur = saved }()

	// Pode começar com <T>
	if p.cur.Lexeme == "<" {
		p.advanceToken()
		for p.cur.Type == lexer.GENERIC && p.cur.Lexeme != ">" {
			p.advanceToken()
			if p.cur.Lexeme == "," {
				p.advanceToken()
			}
		}
		if p.cur.Lexeme != ">" {
			return false
		}
		p.advanceToken()
	}

	// Deve ter um tipo
	if !isTypeKeyword(p.cur.Lexeme) && p.cur.Type != lexer.IDENT {
		return false
	}
	p.advanceToken()

	// Modificadores de tipo (?, *, [])
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

	// Deve ter a palavra "method"
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

	// Generics opcionais: type Car<T>
	var generics []*GenericParam
	if p.cur.Lexeme == "<" {
		generics = p.parseGenericParams()
	}

	// Pode ser union type ou struct type
	typ := p.parseTypeBody()
	if typ == nil {
		return nil
	}

	return &TypeDecl{Name: name, Generics: generics, Type: typ}
}

func (p *Parser) parseTypeBody() Type {
	// Se começar com {, é um struct type
	if p.cur.Lexeme == "{" {
		return p.parseStructType()
	}

	// Caso contrário, é um union type ou tipo simples
	return p.parseType()
}

func (p *Parser) parseStructType() Type {
	p.advanceToken() // consume '{'

	var fields []*FieldDecl

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		field := p.parseField()
		if field != nil {
			fields = append(fields, field)
		} else {
			p.advanceToken()
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructType{Fields: fields}
}
