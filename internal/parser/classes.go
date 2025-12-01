package parser

import "github.com/alpha/internal/lexer"

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

// parseNewExpr parseia new User("name", 10)
func (p *Parser) parseNewExpr() Expr {
	p.advanceToken() // consume 'new'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name after 'new'")
		return nil
	}

	typeName := p.cur.Lexeme
	p.advanceToken()

	// Generics opcionais: new List<int>()
	var typeArgs []Type
	if p.cur.Lexeme == "<" {
		p.advanceToken()
		typeArgs = p.parseTypeArguments()
		if !p.expectAndConsume(">") {
			return nil
		}
	}

	// Argumentos do construtor
	if !p.expectAndConsume("(") {
		return nil
	}

	var args []Expr
	if p.cur.Lexeme != ")" {
		args = p.parseArgumentList()
	}

	if !p.expectAndConsume(")") {
		return nil
	}

	return &NewExpr{
		TypeName: typeName,
		TypeArgs: typeArgs,
		Args:     args,
	}
}

// parseMemberAccess parseia user.name
func (p *Parser) parseMemberAccess(left Expr) Expr {
	p.advanceToken() // consume '.'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected member name after '.'")
		return nil
	}

	member := p.cur.Lexeme
	p.advanceToken()

	return &MemberExpr{Object: left, Member: member}
}
