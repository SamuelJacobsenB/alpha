package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// DECLARAÇÕES DE VARIÁVEIS E CONSTANTES
// ============================

// parseTypedVarDecl processa declarações de variáveis com tipo explícito (usado internamente)
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
		init = p.parseExpression(LOWEST)
		if init == nil {
			p.errorf("expected expression after '='")
			return nil
		}
	}

	return &VarDecl{Name: name, Type: typ, Init: init}
}

// parseVarDecl processa declarações 'var'
func (p *Parser) parseVarDecl() Stmt {
	p.advanceToken() // consome 'var'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after 'var'")
		p.syncTo(";")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	var typ Type
	if p.cur.Lexeme == ":" {
		p.advanceToken()
		typ = p.parseType()
		if typ == nil {
			p.syncTo(";")
			return nil
		}
	}

	init := p.parseOptionalInitializer()
	return &VarDecl{Name: name, Type: typ, Init: init}
}

// parseConstDecl processa declarações 'const'
func (p *Parser) parseConstDecl() Stmt {
	p.advanceToken() // consome 'const'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after 'const'")
		p.syncTo(";")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	if p.cur.Lexeme != "=" {
		p.errorf("expected '=' in const declaration")
		p.syncTo(";")
		return nil
	}

	p.advanceToken()
	init := p.parseExpression(LOWEST)
	if init == nil {
		p.errorf("expected expression for constant value")
		p.syncTo(";")
		return nil
	}

	return &ConstDecl{Name: name, Init: init}
}

// parseOptionalInitializer processa inicializador opcional após '='
func (p *Parser) parseOptionalInitializer() Expr {
	if p.cur.Lexeme != "=" {
		return nil
	}
	p.advanceToken()

	// Otimização: verifica literais comuns antes de parseExpression completo
	switch p.cur.Lexeme {
	case "{":
		return p.parseCollectionLiteral()
	case "[":
		return p.parseArrayLiteral()
	case "&":
		return p.parseReferenceExpr()
	default:
		return p.parseExpression(LOWEST)
	}
}

// ============================
// FUNÇÕES
// ============================

// parseFunctionDecl processa declarações de funções (genéricas ou não)
func (p *Parser) parseFunctionDecl(generic bool) Stmt {
	var generics []*GenericParam
	if generic {
		generics = p.parseGenericParamsList()
		if generics == nil {
			return nil
		}
	}

	returnType := p.parseType()
	if returnType == nil {
		return nil
	}

	if !p.expectAndConsume("function") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected function name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	params := p.parseFunctionParameters()
	if params == nil {
		return nil
	}

	body := p.parseFunctionBody()
	if body == nil {
		return nil
	}

	return &FunctionDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

// parseFunctionParameters processa lista de parâmetros de função
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
		typ := p.parseType()
		if typ == nil {
			p.errorf("expected parameter type")
			return nil
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name")
			return nil
		}

		params = append(params, &Param{Name: p.cur.Lexeme, Type: typ})
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

// parseFunctionBody processa corpo de função/método
func (p *Parser) parseFunctionBody() []Stmt {
	if p.cur.Lexeme != "{" {
		p.errorf("expected '{' to start function/method body")
		return nil
	}

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return body
}

// ============================
// STRUCTS (DADOS)
// ============================

// parseStructDecl processa declarações de struct (com ou sem generics)
func (p *Parser) parseStructDecl(generics []*GenericParam) Stmt {
	if generics == nil {
		generics = p.parseGenericParamsWithPrefix()
	}

	if !p.expectAndConsume("struct") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected struct name")
		p.syncTo("}")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Suporte para generics após o nome (sintaxe alternativa)
	if generics == nil && p.cur.Lexeme == "<" {
		generics = p.parseGenericParamsList()
	}

	if !p.expectAndConsume("{") {
		return nil
	}

	fields := p.parseStructFields()

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructDecl{
		Name:     name,
		Generics: generics,
		Fields:   fields,
	}
}

// parseStructFields processa campos de struct
func (p *Parser) parseStructFields() []*FieldDecl {
	fields := make([]*FieldDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			continue
		}

		isPrivate := false
		switch p.cur.Lexeme {
		case "private":
			isPrivate = true
			p.advanceToken()
		case "public":
			p.advanceToken()
		}

		typ := p.parseType()
		if typ == nil {
			if isPrivate {
				p.errorf("expected type after 'private'")
			} else {
				p.errorf("expected field type in struct, got '%s'", p.cur.Lexeme)
			}
			p.syncStructField()
			continue
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected field name after type, got '%s'", p.cur.Lexeme)
			p.syncStructField()
			continue
		}

		fieldName := p.cur.Lexeme
		p.advanceToken()

		p.consumeOptionalSemicolon()

		fields = append(fields, &FieldDecl{
			Name:      fieldName,
			Type:      typ,
			IsPrivate: isPrivate,
		})
	}

	return fields
}

// syncStructField sincroniza após erro em campo de struct
func (p *Parser) syncStructField() {
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			return
		}
		if p.cur.Lexeme == "private" || p.cur.Lexeme == "public" || isTypeKeyword(p.cur.Lexeme) || p.cur.Type == lexer.IDENT {
			return
		}
		p.advanceToken()
	}
}

// ============================
// IMPLEMENTAÇÕES (COMPORTAMENTO)
// ============================

// parseImplementDecl processa blocos 'implement'
func (p *Parser) parseImplementDecl() Stmt {
	p.advanceToken() // consome 'implement'

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected struct name to implement")
		return nil
	}

	targetName := p.cur.Lexeme
	p.advanceToken()

	if !p.expectAndConsume("{") {
		return nil
	}

	init, methods := p.parseImplementBody()

	if !p.expectAndConsume("}") {
		return nil
	}

	return &ImplDecl{
		TargetName: targetName,
		Init:       init,
		Methods:    methods,
	}
}

// parseImplementBody processa corpo de implementação (init + métodos)
func (p *Parser) parseImplementBody() (*InitDecl, []*MethodDecl) {
	var initDecl *InitDecl
	methods := make([]*MethodDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			continue
		}

		if p.cur.Lexeme == "init" {
			if initDecl != nil {
				p.errorf("multiple init blocks defined")
			}
			initDecl = p.parseInitDecl()
			continue
		}

		method := p.parseMethodDecl()
		if method != nil {
			methods = append(methods, method)
		}
	}

	return initDecl, methods
}

// parseMethodDecl processa declaração de método individual
func (p *Parser) parseMethodDecl() *MethodDecl {
	var generics []*GenericParam
	if p.cur.Lexeme == "generic" && p.nxt.Lexeme == "<" {
		generics = p.parseGenericParamsWithPrefix()
	}

	returnType := p.parseType()
	if returnType == nil {
		p.errorf("expected return type for method")
		p.syncImplMember()
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected method name")
		p.syncImplMember()
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

// parseInitDecl processa declaração de inicializador (init)
func (p *Parser) parseInitDecl() *InitDecl {
	p.advanceToken() // consome 'init'
	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()
	return &InitDecl{Params: params, Body: body}
}

// syncImplMember sincroniza após erro em membro de implementação
func (p *Parser) syncImplMember() {
	p.advanceToken()
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			return
		}
		if p.cur.Lexeme == "init" || p.cur.Lexeme == "generic" {
			return
		}
		if isTypeKeyword(p.cur.Lexeme) {
			return
		}
		p.advanceToken()
	}
}

// ============================
// GENERICS TOP-LEVEL
// ============================

// parseGenericDeclaration processa 'generic<T> ...'
func (p *Parser) parseGenericDeclaration() Stmt {
	generics := p.parseGenericParamsWithPrefix()
	if generics == nil {
		return nil
	}

	switch p.cur.Lexeme {
	case "struct":
		return p.parseStructDecl(generics)
	default:
		returnType := p.parseType()
		if returnType == nil {
			p.errorf("expected struct or return type after generics")
			return nil
		}

		if !p.expectAndConsume("function") {
			p.errorf("expected 'function' keyword")
			return nil
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected function name")
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
}

// parseGenericTopLevel processa atalho '<T> ...'
func (p *Parser) parseGenericTopLevel() Stmt {
	generics := p.parseGenericParamsList()
	if generics == nil {
		return nil
	}

	switch p.cur.Lexeme {
	case "struct":
		return p.parseStructDecl(generics)
	default:
		returnType := p.parseType()
		if returnType == nil {
			return nil
		}
		if !p.expectAndConsume("function") {
			return nil
		}
		if p.cur.Type != lexer.IDENT {
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
}

// ============================
// DECLARAÇÕES DE TIPOS
// ============================

// parseTypeDecl processa declarações 'type' (genéricas ou não)
func (p *Parser) parseTypeDecl() Stmt {
	var generics []*GenericParam
	var hasGenericPrefix bool

	if p.cur.Lexeme == "generic" && p.nxt.Lexeme == "<" {
		generics = p.parseGenericParamsWithPrefix()
		hasGenericPrefix = true
	}

	if !hasGenericPrefix {
		p.advanceToken() // consome 'type'
	} else if !p.expectAndConsume("type") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	if !hasGenericPrefix && p.cur.Lexeme == "<" {
		generics = p.parseGenericParamsList()
	}

	typ := p.parseTypeBody()

	return &TypeDecl{Name: name, Generics: generics, Type: typ}
}

// parseTypeBody processa corpo de declaração de tipo
func (p *Parser) parseTypeBody() Type {
	if p.cur.Lexeme == "{" {
		// Struct anônimo legacy
		p.advanceToken()
		fields := p.parseStructFields()
		if !p.expectAndConsume("}") {
			return nil
		}
		return &StructType{Fields: fields}
	}
	return p.parseType()
}
