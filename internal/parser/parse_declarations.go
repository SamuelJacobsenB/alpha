package parser

import (
	"github.com/alpha/internal/lexer"
)

// ============================
// Declarações de Variáveis/Constantes
// ============================

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

func (p *Parser) parseGenericDeclaration() Stmt {
	generics := p.parseGenericParamsWithPrefix()
	if generics == nil {
		return nil
	}

	// Verifica qual tipo de declaração segue os generics
	switch p.cur.Lexeme {
	case "struct":
		return p.parseStructDecl(generics)
	case "class":
		// Lógica similar a parseClass, mas injetando os generics já parseados
		p.advanceToken() // consume 'class'
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected class name")
			return nil
		}
		name := p.cur.Lexeme
		p.advanceToken()
		if !p.expectAndConsume("{") {
			return nil
		}
		fields, constructor, methods := p.parseClassMembers()
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
	default:
		// Assume que é uma função: generic<T> Tipo function ...
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
}

func (p *Parser) parseGenericFunctionDecl() Stmt {
	generics := p.parseGenericParamsWithPrefix()
	if generics == nil {
		return nil
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

// ============================
// Classes
// ============================

func (p *Parser) parseClass() Stmt {
	// Verificar se é classe genérica
	generics := p.parseGenericParamsWithPrefix()

	if !p.expectAndConsume("class") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected class name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Se não tinha generics no prefixo, verificar se tem após o nome
	if generics == nil && p.cur.Lexeme == "<" {
		generics = p.parseGenericParamsList()
	}

	if !p.expectAndConsume("{") {
		return nil
	}

	fields, constructor, methods := p.parseClassMembers()

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

func (p *Parser) parseClassMembers() ([]*FieldDecl, *ConstructorDecl, []*MethodDecl) {
	fields := make([]*FieldDecl, 0, 4)
	var constructor *ConstructorDecl
	methods := make([]*MethodDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == "constructor" {
			constructor = p.parseConstructor()
			continue
		}

		// Tenta parsear generics (prefixo de método genérico)
		generics := p.parseGenericParamsWithPrefix()
		if generics == nil && p.cur.Lexeme == "<" {
			// Fallback para generics sem prefixo 'generic' se a sintaxe permitir
			generics = p.parseGenericParamsList()
		}

		// Parsear o Tipo (pode ser tipo do campo ou retorno do método)
		typ := p.parseType()
		if typ == nil {
			p.errorf("expected type, 'constructor' or '}' in class definition, got %s", p.cur.Lexeme)
			p.advanceToken() // Evita loop infinito
			continue
		}

		// Verifica se é um método
		if p.cur.Lexeme == "method" {
			p.advanceToken() // consome 'method'

			if p.cur.Type != lexer.IDENT {
				p.errorf("expected method name")
				continue
			}
			methodName := p.cur.Lexeme
			p.advanceToken()

			params := p.parseFunctionParameters()
			body := p.parseFunctionBody()

			methods = append(methods, &MethodDecl{
				Name:       methodName,
				Generics:   generics,
				Params:     params,
				ReturnType: typ,
				Body:       body,
			})
		} else {
			// É um campo
			if generics != nil {
				p.errorf("fields cannot have generic parameters")
			}

			if p.cur.Type != lexer.IDENT {
				p.errorf("expected field name, got %s", p.cur.Lexeme)
				p.advanceToken()
				continue
			}

			fieldName := p.cur.Lexeme
			p.advanceToken()
			p.consumeOptionalSemicolon()

			fields = append(fields, &FieldDecl{Name: fieldName, Type: typ})
		}
	}

	return fields, constructor, methods
}

// parseFieldDecl parseia um campo de classe/struct
// Sintaxe: tipo nome [;]
// NOTA: Mantido para uso em parseTypeBody, mas parseClassMembers/StructMembers agora implementam a lógica inline
func (p *Parser) parseFieldDecl() *FieldDecl {
	// Parsear tipo
	typ := p.parseType()
	if typ == nil {
		p.errorf("expected type for field declaration")
		return nil
	}

	// Verificar nome do campo
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected field name, got %s (%s)", p.cur.Lexeme, p.cur.Type)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Consumir ponto-e-vírgula opcional
	p.consumeOptionalSemicolon()

	return &FieldDecl{Name: name, Type: typ}
}

func (p *Parser) parseConstructor() *ConstructorDecl {
	p.advanceToken() // consume 'constructor'

	params := p.parseFunctionParameters()
	body := p.parseFunctionBody()

	return &ConstructorDecl{Params: params, Body: body}
}

// parseMethod mantido se necessário por outras chamadas, mas a lógica principal foi movida para Members
func (p *Parser) parseMethod() *MethodDecl {
	// Verificar se é método genérico
	generics := p.parseGenericParamsWithPrefix()

	// Se não tinha generics no prefixo, verificar se tem antes do tipo de retorno
	if generics == nil && p.cur.Lexeme == "<" {
		generics = p.parseGenericParamsList()
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

// ============================
// Structs
// ============================

// parseStructDecl parseia uma declaração de struct
// Sintaxe: [generic<T>] struct Name [<T>] { campos... }
func (p *Parser) parseStructDecl(generics []*GenericParam) Stmt {
	// Se generics ainda não foi parseado, tentar parsear
	if generics == nil {
		generics = p.parseGenericParamsWithPrefix()
	}

	if !p.expectAndConsume("struct") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected struct name")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Se não tinha generics no prefixo, verificar se tem após o nome
	if generics == nil && p.cur.Lexeme == "<" {
		generics = p.parseGenericParamsList()
	}

	if !p.expectAndConsume("{") {
		return nil
	}

	// Parsear campos e métodos
	fields, constructor, methods := p.parseStructMembers()

	if !p.expectAndConsume("}") {
		return nil
	}

	// Retorna como ClassDecl pois a AST usa isso para estruturas com métodos
	return &ClassDecl{
		Name:        name,
		Generics:    generics,
		Fields:      fields,
		Constructor: constructor,
		Methods:     methods,
	}
}

// parseStructMembers parseia os membros de uma struct (campos e métodos)
func (p *Parser) parseStructMembers() ([]*FieldDecl, *ConstructorDecl, []*MethodDecl) {
	fields := make([]*FieldDecl, 0, 4)
	var constructor *ConstructorDecl
	methods := make([]*MethodDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		if p.cur.Lexeme == "constructor" {
			constructor = p.parseConstructor()
			continue
		}

		// 1. Tentar parsear Generics (opcional, para métodos)
		generics := p.parseGenericParamsWithPrefix()
		if generics == nil && p.cur.Lexeme == "<" {
			generics = p.parseGenericParamsList()
		}

		// 2. Parsear Tipo Comum (Tipo do Campo ou Retorno do Método)
		typ := p.parseType()
		if typ == nil {
			p.errorf("expected type, 'constructor' or '}' in struct, got %s", p.cur.Lexeme)
			p.advanceToken()
			continue
		}

		// 3. Decidir se é Método ou Campo
		if p.cur.Lexeme == "method" {
			// É um método
			p.advanceToken() // consome 'method'

			if p.cur.Type != lexer.IDENT {
				p.errorf("expected method name")
				continue
			}
			methodName := p.cur.Lexeme
			p.advanceToken()

			params := p.parseFunctionParameters()
			body := p.parseFunctionBody()

			methods = append(methods, &MethodDecl{
				Name:       methodName,
				Generics:   generics,
				Params:     params,
				ReturnType: typ,
				Body:       body,
			})
		} else {
			// É um campo
			if generics != nil {
				p.errorf("struct fields cannot have generic parameters")
			}

			if p.cur.Type != lexer.IDENT {
				p.errorf("expected field name, got %s", p.cur.Lexeme)
				p.advanceToken()
				continue
			}

			fieldName := p.cur.Lexeme
			p.advanceToken()

			p.consumeOptionalSemicolon()

			fields = append(fields, &FieldDecl{Name: fieldName, Type: typ})
		}
	}

	return fields, constructor, methods
}

// ============================
// Declarações de Tipos
// ============================

func (p *Parser) parseTypeDecl() Stmt {
	var generics []*GenericParam
	var hasGenericPrefix bool

	// Verificar se é tipo genérico com prefixo
	if p.cur.Lexeme == "generic" && p.nxt.Lexeme == "<" {
		generics = p.parseGenericParamsWithPrefix()
		if generics == nil {
			return nil
		}
		hasGenericPrefix = true
	}

	if !hasGenericPrefix {
		p.advanceToken() // consume 'type'
	} else if !p.expectAndConsume("type") {
		return nil
	}

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected type name after 'type'")
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	// Verificar parâmetros genéricos após o nome
	if !hasGenericPrefix && p.cur.Lexeme == "<" {
		generics = p.parseGenericParamsList()
	}

	typ := p.parseTypeBody()
	if typ == nil {
		return nil
	}

	return &TypeDecl{Name: name, Generics: generics, Type: typ}
}

func (p *Parser) parseTypeBody() Type {
	if p.cur.Lexeme != "{" {
		return p.parseType()
	}

	p.advanceToken() // consume '{'
	fields := make([]*FieldDecl, 0, 4)

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		field := p.parseFieldDecl()
		if field != nil {
			fields = append(fields, field)
		}
	}

	if !p.expectAndConsume("}") {
		return nil
	}

	return &StructType{Fields: fields}
}

// ============================
// Generic Top Level
// ============================

// parseGenericTopLevel lida com declarações que começam com <T>
func (p *Parser) parseGenericTopLevel() Stmt {
	generics := p.parseGenericParamsList() // Consome <T>
	if generics == nil {
		return nil
	}

	switch p.cur.Lexeme {
	case "struct":
		return p.parseStructDecl(generics)
	case "class":
		// Salvar estado e tentar parsear como classe
		p.advanceToken() // consume 'class'

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected class name")
			return nil
		}

		name := p.cur.Lexeme
		p.advanceToken()

		if !p.expectAndConsume("{") {
			return nil
		}

		fields, constructor, methods := p.parseClassMembers()

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
	default:
		// Pode ser uma função: <T> int function...
		returnType := p.parseType()
		if returnType == nil {
			return nil
		}

		if !p.expectAndConsume("function") {
			p.errorf("expected 'function', 'struct', or 'class' after generic parameters")
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
