package parser

import (
	"fmt"

	"github.com/alpha/internal/lexer"
)

func (p *Parser) parseTopLevel() Stmt {
	fmt.Printf("parseTopLevel: cur=%q (type=%v) nxt=%q\n", p.cur.Lexeme, p.cur.Type, p.nxt.Lexeme)

	if p.cur.Type == lexer.EOF {
		return nil
	}

	// CORREÇÃO MELHORADA: Detecção de funções genéricas
	if p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme) {
		// Lookahead mais inteligente para funções genéricas
		if p.nxt.Lexeme == "function" || (p.nxt.Lexeme == "[" && p.peekNextAfterBrackets() == "function") {
			fmt.Printf("parseTopLevel: detected FUNCTION declaration (possibly generic)\n")
			return p.parseFunctionDecl()
		}

		// Se não for função, é declaração de variável tipada
		return p.parseTypedVarDecl()
	}

	if p.cur.Type == lexer.KEYWORD {
		switch p.cur.Lexeme {
		case "var":
			fmt.Println("parseTopLevel: detected VAR")
			return p.parseVarDecl()
		case "const":
			fmt.Println("parseTopLevel: detected CONST")
			return p.parseConstDecl()
		case "function":
			// CORREÇÃO: Não permitir função anônima no top-level
			p.errorf("anonymous functions are not allowed at top level")
			p.advanceToken() // skip 'function'
			return nil
		case "if":
			fmt.Println("parseTopLevel: detected IF")
			return p.parseIf()
		case "while":
			fmt.Println("parseTopLevel: detected WHILE")
			return p.parseWhile()
		case "for":
			fmt.Println("parseTopLevel: detected FOR")
			return p.parseFor()
		case "return":
			fmt.Println("parseTopLevel: detected RETURN")
			return p.parseReturn()
		default:
			p.advanceToken()
			return nil
		}
	}

	if p.cur.Lexeme == "{" {
		fmt.Println("parseTopLevel: detected BLOCK")
		return &BlockStmt{Body: p.parseBlockLike()}
	}

	if p.cur.Lexeme == "}" {
		p.advanceToken()
		return nil
	}

	if p.canStartExpression() {
		fmt.Println("parseTopLevel: parsing expression statement")
		return p.parseExprStmt()
	}

	fmt.Printf("parseTopLevel: unrecognized token %q, advancing\n", p.cur.Lexeme)
	p.advanceToken()
	return nil
}

// CORREÇÃO: Nova função auxiliar para lookahead após colchetes
func (p *Parser) peekNextAfterBrackets() string {
	saveCur := p.cur
	saveNxt := p.nxt

	// Avançar até encontrar o fechamento de ]
	for p.cur.Lexeme != "]" && p.cur.Type != lexer.EOF {
		p.advanceToken()
	}

	if p.cur.Lexeme == "]" {
		p.advanceToken() // consume ]
	}

	next := p.cur.Lexeme

	// Restaurar estado
	p.cur = saveCur
	p.nxt = saveNxt

	return next
}

// Nova função auxiliar
func (p *Parser) canStartExpression() bool {
	switch p.cur.Type {
	case lexer.IDENT, lexer.INT, lexer.FLOAT, lexer.STRING:
		return true
	case lexer.KEYWORD:
		return p.cur.Lexeme == "true" || p.cur.Lexeme == "false" || p.cur.Lexeme == "null"
	case lexer.OP:
		return p.cur.Lexeme == "-" || p.cur.Lexeme == "!" || p.cur.Lexeme == "+" ||
			p.cur.Lexeme == "(" || p.cur.Lexeme == "{"
	default:
		return false
	}
}

func (p *Parser) parseTypedVarDecl() Stmt {
	fmt.Printf("parseTypedVarDecl: starting\n")

	// Usar parseType em vez de parsear manualmente
	typ := p.parseType()
	if typ == nil {
		return nil
	}

	fmt.Printf("parseTypedVarDecl: parsed type %T, cur=%q\n", typ, p.cur.Lexeme)

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after type at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()
	fmt.Printf("parseTypedVarDecl: after ident, cur=%q\n", p.cur.Lexeme)

	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		fmt.Printf("parseTypedVarDecl: after '=', cur=%q\n", p.cur.Lexeme)
		init = p.parseExpression(LOWEST)
	}

	fmt.Printf("parseTypedVarDecl: completed %s, cur=%q\n", name, p.cur.Lexeme)
	return &VarDecl{Name: name, Type: typ, Init: init}
}

func (p *Parser) parseVarDecl() Stmt {
	p.advanceToken()

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier after var at %d:%d", p.cur.Line, p.cur.Col)
		return nil
	}

	name := p.cur.Lexeme
	p.advanceToken()

	var init Expr
	if p.cur.Lexeme == "=" {
		p.advanceToken()
		init = p.parseExpression(LOWEST)
	}

	return &VarDecl{Name: name, Init: init}
}

func (p *Parser) parseConstDecl() Stmt {
	p.advanceToken()

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

func (p *Parser) parseExprStmt() Stmt {
	ex := p.parseExpression(LOWEST)
	return &ExprStmt{Expr: ex}
}

func (p *Parser) parseReturn() Stmt {
	p.advanceToken() // consume 'return'

	if p.cur.Type == lexer.EOF || p.cur.Lexeme == "}" || p.cur.Lexeme == ";" {
		return &ReturnStmt{Value: nil}
	}

	value := p.parseExpression(LOWEST)
	return &ReturnStmt{Value: value}
}

func (p *Parser) parseIf() Stmt {
	p.advanceToken()

	hasParen := false
	if p.cur.Lexeme == "(" {
		hasParen = true
		p.advanceToken()
	}

	cond := p.parseExpression(LOWEST)
	if cond == nil {
		p.errorf("invalid condition in if statement")
		return nil
	}

	if hasParen {
		if p.cur.Lexeme == ")" {
			p.advanceToken()
		} else {
			p.errorf("expected ')' after if condition at %d:%d", p.cur.Line, p.cur.Col)
			return nil
		}
	}

	thenBlock := p.parseBlockLike()

	var elseBlock []Stmt
	if p.cur.Lexeme == "else" {
		p.advanceToken()
		elseBlock = p.parseBlockLike()
	}

	return &IfStmt{Cond: cond, Then: thenBlock, Else: elseBlock}
}

func (p *Parser) parseWhile() Stmt {
	p.advanceToken()

	if p.cur.Lexeme == "(" {
		p.advanceToken()
	}

	cond := p.parseExpression(LOWEST)

	if p.cur.Lexeme == ")" {
		p.advanceToken()
	}

	body := p.parseBlockLike()
	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseFor() Stmt {
	fmt.Printf("parseFor: starting, cur=%q nxt=%q\n", p.cur.Lexeme, p.nxt.Lexeme)

	// CORREÇÃO SIMPLIFICADA: Detectar for...in de forma mais robusta
	if p.nxt.Lexeme == "(" {
		// Salvar estado para lookahead
		saveCur := p.cur
		saveNxt := p.nxt

		// Fazer lookahead limitado
		p.advanceToken() // cur = "("
		p.advanceToken() // cur = primeiro token dentro dos parênteses

		isForIn := false
		steps := 0

		// Procurar por padrão: IDENT [',' IDENT] 'in'
		for steps < 5 && p.cur.Type != lexer.EOF && p.cur.Lexeme != ")" {
			if p.cur.Lexeme == "in" {
				isForIn = true
				break
			}
			p.advanceToken()
			steps++
		}

		// Restaurar estado
		p.cur = saveCur
		p.nxt = saveNxt

		if isForIn {
			fmt.Println("parseFor: detected for...in loop")
			return p.parseForIn()
		}
	}

	fmt.Println("parseFor: detected traditional for loop")
	return p.parseForTraditional()
}

func (p *Parser) parseFunctionDecl() Stmt {
	fmt.Printf("parseFunctionDecl: starting with cur=%q\n", p.cur.Lexeme)

	// CORREÇÃO: Parsear tipo de retorno (pode ser void)
	returnType := p.parseType()
	if returnType == nil {
		p.errorf("expected return type for function")
		return nil
	}

	// Palavra-chave 'function'
	if p.cur.Lexeme != "function" {
		p.errorf("expected 'function' keyword after return type, got %q", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume 'function'

	// Nome da função
	if p.cur.Type != lexer.IDENT {
		p.errorf("expected function name, got %q", p.cur.Lexeme)
		return nil
	}
	name := p.cur.Lexeme
	p.advanceToken()

	// CORREÇÃO: Parâmetros genéricos [T] - verificar se existe '['
	var generics []*GenericParam
	if p.cur.Lexeme == "[" {
		generics = p.parseGenericParams()
		// Se parseGenericParams retornar nil, continuar mesmo assim
	}

	// Parâmetros da função
	if p.cur.Lexeme != "(" {
		p.errorf("expected '(' after function name, got %q", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume '('

	params := p.parseFunctionParams()

	if p.cur.Lexeme != ")" {
		p.errorf("expected ')' after function parameters, got %q", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume ')'

	// Corpo da função
	body := p.parseBlockLike()
	if body == nil {
		p.errorf("expected function body")
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

// CORREÇÃO: parseGenericParams mais robusto
func (p *Parser) parseGenericParams() []*GenericParam {
	fmt.Printf("parseGenericParams: starting at %q\n", p.cur.Lexeme)
	p.advanceToken() // consume '['

	var generics []*GenericParam

	for p.cur.Lexeme != "]" && p.cur.Type != lexer.EOF {
		if p.cur.Type == lexer.IDENT {
			generics = append(generics, &GenericParam{Name: p.cur.Lexeme})
			p.advanceToken()
		} else {
			p.errorf("expected generic parameter name, got %q", p.cur.Lexeme)
			break
		}

		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "]" {
			p.errorf("expected ',' or ']' in generic parameters")
			break
		}
	}

	if p.cur.Lexeme == "]" {
		p.advanceToken() // consume ']'
	} else {
		p.errorf("expected ']' after generic parameters")
	}

	fmt.Printf("parseGenericParams: found %d generic parameters\n", len(generics))
	return generics
}

// parseFunctionParams parseia parâmetros de função
func (p *Parser) parseFunctionParams() []*Param {
	fmt.Printf("parseFunctionParams: starting, cur=%q\n", p.cur.Lexeme)
	var params []*Param

	for p.cur.Lexeme != ")" && p.cur.Type != lexer.EOF {
		// Tipo do parâmetro
		paramType := p.parseType()
		if paramType == nil {
			p.errorf("expected parameter type")
			return nil
		}

		// Nome do parâmetro
		if p.cur.Type != lexer.IDENT {
			p.errorf("expected parameter name")
			return nil
		}
		paramName := p.cur.Lexeme
		p.advanceToken()

		params = append(params, &Param{
			Name: paramName,
			Type: paramType,
		})

		// Verificar se há mais parâmetros
		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != ")" {
			p.errorf("expected ',' or ')' after parameter")
			return nil
		}
	}

	fmt.Printf("parseFunctionParams: completed with %d params\n", len(params))
	return params
}

func (p *Parser) parseFunctionExpr() Expr {
	fmt.Printf("parseFunctionExpr: starting at %q\n", p.cur.Lexeme)

	// Tipo de retorno
	returnType := p.parseType()
	if returnType == nil {
		return nil
	}

	// Palavra-chave 'function'
	if p.cur.Lexeme != "function" {
		p.errorf("expected 'function' keyword, got %q", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume 'function'

	// Parâmetros genéricos [T] (opcional)
	var generics []*GenericParam
	if p.cur.Lexeme == "[" {
		generics = p.parseGenericParams()
	}

	// Parâmetros
	if p.cur.Lexeme != "(" {
		p.errorf("expected '(' for function parameters, got %q", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume '('

	params := p.parseFunctionParams()

	if p.cur.Lexeme != ")" {
		p.errorf("expected ')' after function parameters, got %q", p.cur.Lexeme)
		return nil
	}
	p.advanceToken() // consume ')'

	// Corpo da função
	body := p.parseBlockLike()
	if body == nil {
		return nil
	}

	return &FunctionExpr{
		Generics:   generics,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseForTraditional() Stmt {
	fmt.Printf("parseForTraditional: starting\n")
	p.advanceToken() // consome 'for'

	var init Stmt
	var cond Expr
	var post Stmt

	if p.cur.Lexeme == "(" {
		p.advanceToken() // consome '('

		// CORREÇÃO: Limitar número de passos para evitar loops
		maxSteps := 20
		steps := 0

		// Parsear init
		if p.cur.Lexeme != ";" && steps < maxSteps {
			if p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme) {
				init = p.parseTypedVarDecl()
			} else if p.cur.Lexeme == "var" {
				init = p.parseVarDecl()
			} else {
				init = p.parseExprStmt()
			}
			steps++
		}

		if p.cur.Lexeme == ";" {
			p.advanceToken() // consome ';'
			steps++
		} else if steps < maxSteps {
			p.errorf("expected ';' after for loop initializer")
			return nil
		}

		// Parsear condição
		if p.cur.Lexeme != ";" && steps < maxSteps {
			cond = p.parseExpression(LOWEST)
			steps++
		}

		if p.cur.Lexeme == ";" {
			p.advanceToken() // consome ';'
			steps++
		} else if steps < maxSteps {
			p.errorf("expected ';' after for loop condition")
			return nil
		}

		// Parsear post
		if p.cur.Lexeme != ")" && steps < maxSteps {
			postExpr := p.parseExpression(LOWEST)
			if postExpr != nil {
				post = &ExprStmt{Expr: postExpr}
			}
			steps++
		}

		if p.cur.Lexeme == ")" {
			p.advanceToken() // consome ')'
		} else if steps < maxSteps {
			p.errorf("expected ')' after for loop post statement, got %q", p.cur.Lexeme)
			return nil
		}
	}

	// Parse do corpo
	body := p.parseBlockLike()

	fmt.Printf("parseForTraditional: completed\n")
	return &ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}
}

func (p *Parser) parseForIn() Stmt {
	fmt.Printf("parseForIn: starting\n")
	p.advanceToken() // consome 'for'

	// Verificar se há '('
	if p.cur.Lexeme != "(" {
		p.errorf("expected '(' after 'for' in for...in loop")
		return nil
	}
	p.advanceToken() // consome '('

	// Parse dos identificadores (pode ser um ou dois)
	var indexIdent *Identifier
	var itemIdent *Identifier

	if p.cur.Type != lexer.IDENT {
		p.errorf("expected identifier in for...in loop")
		return nil
	}

	firstIdent := &Identifier{Name: p.cur.Lexeme}
	p.advanceToken()

	// Verificar se há vírgula (caso com índice e item)
	if p.cur.Lexeme == "," {
		p.advanceToken() // consome ','

		if p.cur.Type != lexer.IDENT {
			p.errorf("expected second identifier after ',' in for...in loop")
			return nil
		}

		indexIdent = firstIdent
		itemIdent = &Identifier{Name: p.cur.Lexeme}
		p.advanceToken()
	} else {
		// Apenas um identificador (apenas item)
		indexIdent = nil
		itemIdent = firstIdent
	}

	// Verificar 'in'
	if p.cur.Lexeme != "in" {
		p.errorf("expected 'in' in for...in loop")
		return nil
	}
	p.advanceToken() // consome 'in'

	// Parse da expressão iterável
	iterable := p.parseExpression(LOWEST)
	if iterable == nil {
		p.errorf("invalid iterable expression in for...in loop")
		return nil
	}

	// Verificar ')'
	if p.cur.Lexeme != ")" {
		p.errorf("expected ')' after for...in loop")
		return nil
	}
	p.advanceToken() // consome ')'

	// Parse do corpo do loop
	body := p.parseBlockLike()

	fmt.Printf("parseForIn: completed %s in %T\n", itemIdent.Name, iterable)
	return &ForInStmt{
		Index:    indexIdent,
		Item:     itemIdent,
		Iterable: iterable,
		Body:     body,
	}
}

func (p *Parser) parseBlockLike() []Stmt {
	// CORREÇÃO: Verificação para single statement
	if p.cur.Lexeme != "{" {
		stmt := p.parseTopLevel()
		if stmt != nil {
			return []Stmt{stmt}
		}
		return nil
	}

	p.advanceToken() // consume '{'
	fmt.Printf("parseBlockLike: entered block, cur=%q\n", p.cur.Lexeme)

	var stmts []Stmt

	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		// CORREÇÃO: Guardar token anterior para detectar se estamos presos
		previousToken := p.cur.Lexeme

		stmt := p.parseTopLevel()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}

		// CORREÇÃO: Se não parseamos nada e o token não mudou, avançar
		if stmt == nil && p.cur.Lexeme == previousToken && p.cur.Lexeme != "}" {
			p.advanceToken()
		}

		// Condição de saída
		if p.cur.Lexeme == "}" || p.cur.Type == lexer.EOF {
			break
		}
	}

	if p.cur.Lexeme == "}" {
		p.advanceToken() // consume '}'
	} else if p.cur.Type != lexer.EOF {
		p.errorf("expected '}' to close block, got %q", p.cur.Lexeme)
	}

	fmt.Printf("parseBlockLike: exited block with %d statements\n", len(stmts))
	return stmts
}

func (p *Parser) parseType() Type {
	fmt.Printf("parseType: cur=%q\n", p.cur.Lexeme)

	// CORREÇÃO: Verificar se é realmente um tipo
	if p.cur.Type == lexer.KEYWORD && isTypeKeyword(p.cur.Lexeme) {
		typeName := p.cur.Lexeme
		p.advanceToken()

		// Verificar se é array type
		if p.cur.Lexeme == "[" {
			return p.parseArrayType(&PrimitiveType{Name: typeName})
		}

		return &PrimitiveType{Name: typeName}
	}

	// CORREÇÃO: Não tentar parsear 'function' como tipo
	if p.cur.Lexeme == "function" {
		p.errorf("unexpected 'function' keyword in type position")
		return nil
	}

	p.errorf("expected type, got %q", p.cur.Lexeme)
	return nil
}

func (p *Parser) parseArrayType(elementType Type) Type {
	fmt.Printf("parseArrayType: starting\n")
	p.advanceToken() // consume '['

	var size Expr
	if p.cur.Lexeme != "]" {
		size = p.parseExpression(LOWEST)
	}

	if p.cur.Lexeme != "]" {
		p.errorf("expected ']' in array type")
		return nil
	}
	p.advanceToken() // consume ']'

	return &ArrayType{
		ElementType: elementType,
		Size:        size,
	}
}

func (p *Parser) parseArrayLiteral() Expr {
	fmt.Printf("parseArrayLiteral: starting\n")
	p.advanceToken() // consume '{'

	elements := []Expr{}
	for p.cur.Lexeme != "}" && p.cur.Type != lexer.EOF {
		elem := p.parseExpression(LOWEST)
		if elem == nil {
			return nil
		}
		elements = append(elements, elem)

		if p.cur.Lexeme == "," {
			p.advanceToken()
		} else if p.cur.Lexeme != "}" {
			p.errorf("expected ',' or '}' in array literal")
			return nil
		}
	}

	if p.cur.Lexeme != "}" {
		p.errorf("expected '}' after array literal")
		return nil
	}
	p.advanceToken() // consume '}'

	return &ArrayLiteral{Elements: elements}
}

func (p *Parser) parseIndexExpression(left Expr) Expr {
	fmt.Printf("parseIndexExpression: starting\n")
	p.advanceToken() // consume '['

	index := p.parseExpression(LOWEST)
	if index == nil {
		return nil
	}

	if p.cur.Lexeme != "]" {
		p.errorf("expected ']' after index expression")
		return nil
	}
	p.advanceToken() // consume ']'

	return &IndexExpr{Array: left, Index: index}
}
