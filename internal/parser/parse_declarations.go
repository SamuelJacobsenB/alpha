package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// Declarações de Variáveis e Constantes
// ============================

func (p *Parser) parseTypedVarDecl() Stmt {
	typ := p.parseType()
	// Se parseType falhar ou não houver identificador a seguir, retorna nil imediatamente
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
			// Não sincronizamos aqui pois parseTypedVarDecl é geralmente especulativo
			return nil
		}
	}

	return &VarDecl{Name: name, Type: typ, Init: init}
}

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

func (p *Parser) parseOptionalInitializer() Expr {
	if p.cur.Lexeme != "=" {
		return nil
	}
	p.advanceToken()

	// Otimização: verifica literais comuns antes de chamar parseExpression completo
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
// Funções
// ============================

func (p *Parser) parseFunctionDecl(generic bool) Stmt {
	var generics []*GenericParam
	if generic {
		generics = p.parseGenericParamsList()
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
	body := p.parseFunctionBody()

	return &FunctionDecl{
		Name:       name,
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

// ============================
// Structs (Dados)
// ============================

// parseStructDecl lida com 'struct Name { ... }' e 'generic<T> struct Name { ... }'
func (p *Parser) parseStructDecl(generics []*GenericParam) Stmt {
	// Se generics não foram passados, tenta ler o prefixo
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

func (p *Parser) parseStructFields() []*FieldDecl {
	fields := make([]*FieldDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		// Ignorar pontos e vírgulas soltos
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			continue
		}

		isPrivate := false
		if p.cur.Lexeme == "private" {
			isPrivate = true
			p.advanceToken()
		} else if p.cur.Lexeme == "public" {
			// Apenas consome 'public', pois é o padrão (ou tratado externamente)
			p.advanceToken()
		}

		// Parse o tipo do campo
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

		// Parse o nome do campo
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected field name after type, got '%s'", p.cur.Lexeme)
			p.syncStructField()
			continue
		}

		fieldName := p.cur.Lexeme
		p.advanceToken()

		// Consumir ponto e vírgula opcional
		p.consumeOptionalSemicolon()

		fields = append(fields, &FieldDecl{
			Name:      fieldName,
			Type:      typ,
			IsPrivate: isPrivate,
		})
	}

	return fields
}

func (p *Parser) syncStructField() {
	// Avança até encontrar próximo campo ou fim do struct
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			return
		}
		// Verifica se é início de próximo campo
		if p.cur.Lexeme == "private" || p.cur.Lexeme == "public" || isTypeKeyword(p.cur.Lexeme) || p.cur.Type == lexer.IDENT {
			return
		}
		p.advanceToken()
	}
}

// ============================
// Implementações (Comportamento)
// ============================

// parseImplementDecl lida com 'implement Name { ... }'
func (p *Parser) parseImplementDecl() Stmt {
	p.advanceToken() // consume 'implement'

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

func (p *Parser) parseImplementBody() (*InitDecl, []*MethodDecl) {
	var initDecl *InitDecl
	methods := make([]*MethodDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		// Ignorar pontos e vírgulas
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			continue
		}

		// Parse Init (Construtor)
		if p.cur.Lexeme == "init" {
			if initDecl != nil {
				p.errorf("multiple init blocks defined")
			}
			initDecl = p.parseInitDecl()
			continue
		}

		// Parse Métodos
		// Verifica se tem prefixo 'generic' para métodos genéricos
		var memberGenerics []*GenericParam
		if p.cur.Lexeme == "generic" && p.nxt.Lexeme == "<" {
			memberGenerics = p.parseGenericParamsWithPrefix()
		}

		// Parse Tipo de Retorno
		returnType := p.parseType()
		if returnType == nil {
			p.errorf("expected return type for method")
			p.syncClassMember()
			continue
		}

		// Expectativa: Identificador do método
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected method name")
			p.syncClassMember()
			continue
		}

		methodName := p.cur.Lexeme
		p.advanceToken()

		params := p.parseFunctionParameters()
		body := p.parseFunctionBody()

		methods = append(methods, &MethodDecl{
			Name:       methodName,
			Generics:   memberGenerics,
			Params:     params,
			ReturnType: returnType,
			Body:       body,
		})
	}

	return initDecl, methods
}

func (p *Parser) parseInitDecl() *InitDecl {
	p.advanceToken() // consume 'init'

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &InitDecl{Params: params, Body: body}
}

// syncClassMember (agora syncImplMember) avança o parser
func (p *Parser) syncClassMember() {
	p.advanceToken()
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == ";" {
			p.advanceToken()
			return
		}
		if p.cur.Lexeme == "init" || p.cur.Lexeme == "generic" {
			return
		}
		// Heurística de tipo
		if isTypeKeyword(p.cur.Lexeme) {
			return
		}
		p.advanceToken()
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
		typ := p.parseType()
		if typ == nil {
			p.errorf("expected parameter type")
			p.syncTo(")")
			return nil
		}

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name")
			p.syncTo(")")
			return nil
		}

		params = append(params, &Param{Name: p.cur.Lexeme, Type: typ})
		p.advanceToken()

		if p.cur.Lexeme == ")" {
			p.advanceToken()
			break
		}

		if !p.expectAndConsume(",") {
			p.syncTo(")")
			return nil
		}
	}

	return params
}

func (p *Parser) parseFunctionBody() []Stmt {
	if p.cur.Lexeme != "{" {
		p.errorf("expected '{' to start function/method body")
		return nil
	}
	return p.parseBlock()
}

// ============================
// Tratamento de Generics Top-Level
// ============================

// parseGenericDeclaration lida com 'generic<T> ...'
func (p *Parser) parseGenericDeclaration() Stmt {
	generics := p.parseGenericParamsWithPrefix()
	if generics == nil {
		return nil
	}

	switch p.cur.Lexeme {
	case "struct":
		return p.parseStructDecl(generics)

	default:
		// Assume Função Genérica: generic<T> Type function ...
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

// parseGenericTopLevel lida com atalho '<T> ...'
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
		return &FunctionDecl{Name: name, Generics: generics, Params: params, ReturnType: returnType, Body: body}
	}
}

// ============================
// Declarações de Tipos (Alias/Structs antigos)
// ============================

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

func (p *Parser) parseTypeBody() Type {
	if p.cur.Lexeme == "{" {
		// Modo legacy para type alias de struct anonima, se necessário
		// Reutiliza parseStructFields mas retorna StructType
		p.advanceToken()
		fields := p.parseStructFields()
		if !p.expectAndConsume("}") {
			return nil
		}
		return &StructType{Fields: fields}
	}
	return p.parseType()
}
