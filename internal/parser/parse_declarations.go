package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// DECLARAÇÕES DE VARIÁVEIS E CONSTANTES
// ============================

// parseVarDecl processa declarações 'var' (simples ou múltiplas)
func (p *Parser) parseVarDecl() Stmt {
	p.advanceToken() // consome 'var'

	// Parse uma lista de identificadores
	var names []string
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after 'var'")
		p.syncTo(";")
		return nil
	}
	names = append(names, p.cur.Lexeme)
	p.advanceToken()

	// Parse identificadores adicionais separados por vírgula
	for p.cur.Lexeme == "," {
		p.advanceToken() // consome ','

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected identifier after ',' in variable declaration")
			p.syncTo(";")
			return nil
		}
		names = append(names, p.cur.Lexeme)
		p.advanceToken()
	}

	// Tipo opcional (após ':')
	var typ Type
	if p.cur.Lexeme == ":" {
		p.advanceToken()
		typ = p.parseType()
		if typ == nil {
			p.syncTo(";")
			return nil
		}
	}

	// Inicializador obrigatório
	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		init = p.parseExpression(LOWEST)
		if init == nil {
			p.errorf("expected expression after '=' in variable declaration")
			p.syncTo(";")
			return nil
		}
	} else {
		// Para múltiplas variáveis, o inicializador é obrigatório
		if len(names) > 1 {
			p.errorf("multiple variables declaration must have initializer")
			p.syncTo(";")
			return nil
		}
		// Para variável única, o inicializador é opcional
	}

	// Retorna declaração apropriada
	if len(names) == 1 {
		return &VarDecl{Name: names[0], Type: typ, Init: init}
	} else {
		return &MultiVarDecl{Names: names, Type: typ, Init: init}
	}
}

// parseConstDecl processa declarações 'const' (simples ou múltiplas)
func (p *Parser) parseConstDecl() Stmt {
	p.advanceToken() // consome 'const'

	// Parse uma lista de identificadores
	var names []string
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after 'const'")
		p.syncTo(";")
		return nil
	}
	names = append(names, p.cur.Lexeme)
	p.advanceToken()

	// Parse identificadores adicionais separados por vírgula
	for p.cur.Lexeme == "," {
		p.advanceToken() // consome ','

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected identifier after ',' in constant declaration")
			p.syncTo(";")
			return nil
		}
		names = append(names, p.cur.Lexeme)
		p.advanceToken()
	}

	// Inicializador obrigatório
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

	// Retorna declaração apropriada
	if len(names) == 1 {
		return &ConstDecl{Name: names[0], Init: init}
	} else {
		return &MultiConstDecl{Names: names, Init: init}
	}
}

// parseTypedVarDecl processa declarações de variáveis com tipo explícito (apenas simples)
func (p *Parser) parseTypedVarDecl() Stmt {
	// Salva estado para backtracking se falhar
	savedCur, savedNxt := p.cur, p.nxt

	typ := p.parseType()
	if typ == nil {
		return nil
	}

	// Parse uma lista de identificadores (apenas um para typed declaration)
	var names []string
	if p.cur.Type != lexer.IDENT {
		p.cur, p.nxt = savedCur, savedNxt // Restaura
		return nil
	}
	names = append(names, p.cur.Lexeme)
	p.advanceToken()

	// Verifica se há mais identificadores (não permitido em typed declaration)
	if p.cur.Lexeme == "," {
		p.cur, p.nxt = savedCur, savedNxt // Restaura
		return nil
	}

	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		init = p.parseExpression(LOWEST)
		if init == nil {
			p.errorf("expected expression after '='")
			return nil
		}
	}

	return &VarDecl{Name: names[0], Type: typ, Init: init}
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

	returnTypes := p.parseReturnTypeList()
	if returnTypes == nil {
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
		Name:        name,
		Generics:    generics,
		Params:      params,
		ReturnTypes: returnTypes,
		Body:        body,
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
		// Primeiro: parse do tipo
		typ := p.parseType()
		if typ == nil {
			p.errorf("expected parameter type")
			return nil
		}

		// Segundo: parse do nome do parâmetro
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name")
			return nil
		}
		name := p.cur.Lexeme
		p.advanceToken()

		params = append(params, &Param{Name: name, Type: typ})

		// Se tem vírgula, continua para próximo parâmetro
		if p.cur.Lexeme == "," {
			p.advanceToken()
			continue
		}

		// Se não tem vírgula, terminou
		if p.cur.Lexeme == ")" {
			p.advanceToken()
			break
		}

		p.errorf("expected ',' or ')' in parameter list")
		return nil
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
		} else {
			// Sincroniza se falhar no método
			p.syncImplMember()
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

	returnTypes := p.parseReturnTypeList()
	if returnTypes == nil {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		// Não consome se falhar, deixa sync lidar
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &MethodDecl{
		Name:        name,
		Generics:    generics,
		Params:      params,
		ReturnTypes: returnTypes,
		Body:        body,
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
	if p.cur.Type == lexer.EOF || p.cur.Lexeme == "}" {
		return
	}
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
		returnTypes := p.parseReturnTypeList()
		if returnTypes == nil {
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
			Name:        name,
			Generics:    generics,
			Params:      params,
			ReturnTypes: returnTypes,
			Body:        body,
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
		returnTypes := p.parseReturnTypeList()
		if returnTypes == nil {
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
			Name:        name,
			Generics:    generics,
			Params:      params,
			ReturnTypes: returnTypes,
			Body:        body,
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
