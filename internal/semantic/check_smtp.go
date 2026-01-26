package semantic

import (
	"fmt"
	"strings"

	"github.com/alpha/internal/parser"
)

func (c *Checker) checkStmt(stmt parser.Stmt) {
	switch s := stmt.(type) {

	case *parser.ConstDecl:
		c.checkConstDecl(s)

	case *parser.VarDecl:
		c.checkVarDecl(s)

	case *parser.MultiVarDecl:
		c.checkMultiVarDecl(s)

	case *parser.MultiConstDecl:
		c.checkMultiConstDecl(s)

	case *parser.FunctionDecl:
		c.checkFunctionDecl(s)

	case *parser.StructDecl:
		c.checkStructDecl(s)

	case *parser.TypeDecl:
		c.checkTypeDecl(s)

	case *parser.ImplDecl:
		c.checkImplDecl(s)

	case *parser.BlockStmt:
		c.checkBlock(s.Body)

	case *parser.IfStmt:
		condType := c.checkExpr(s.Cond)
		if !c.isConditionableType(condType) {
			c.reportError(0, 0, fmt.Sprintf("Condition in 'if' must be boolean or nullable, got %s", StringifyType(condType)))
		}
		c.checkBlockScope(s.Then)
		if s.Else != nil {
			c.checkBlockScope(s.Else)
		}

	case *parser.WhileStmt:
		condType := c.checkExpr(s.Cond)
		if !c.isConditionableType(condType) {
			c.reportError(0, 0, "Condition in 'while' must be boolean or nullable")
		}
		prevLoop := c.inLoop
		c.inLoop = true
		c.checkBlockScope(s.Body)
		c.inLoop = prevLoop

	case *parser.DoWhileStmt:
		prevLoop := c.inLoop
		c.inLoop = true
		c.checkBlockScope(s.Body)
		c.inLoop = prevLoop

		condType := c.checkExpr(s.Cond)
		if !c.isBooleanType(condType) {
			c.reportError(0, 0, "Condition in 'do-while' must be boolean")
		}

	case *parser.ForStmt:
		c.enterScope() // For cria escopo para o Init
		if s.Init != nil {
			c.checkStmt(s.Init)
		}
		if s.Cond != nil {
			condType := c.checkExpr(s.Cond)
			if !c.isBooleanType(condType) {
				c.reportError(0, 0, "Condition in 'for' must be boolean")
			}
		}
		if s.Post != nil {
			c.checkStmt(s.Post)
		}

		prevLoop := c.inLoop
		c.inLoop = true
		for _, bodyStmt := range s.Body {
			c.checkStmt(bodyStmt)
		}
		c.inLoop = prevLoop

		c.exitScope()

	case *parser.SwitchStmt:
		c.checkSwitchStmt(s)

	case *parser.ReturnStmt:
		if c.currentFuncReturnType == nil {
			c.reportError(0, 0, "Return statement outside of function")
			return
		}

		// Se não há valores de retorno
		if s.Values == nil || len(s.Values) == 0 {
			// Verificar se a função é void
			if c.currentFuncReturnType != nil &&
				StringifyType(c.currentFuncReturnType) != "void" {
				c.reportError(0, 0, "Non-void function must return a value")
			}
			return
		}

		// Verificar se o tipo de retorno da função é MultiValueType
		if multiRet, ok := c.currentFuncReturnType.(*MultiValueType); ok {
			// Função retorna múltiplos valores

			// Verificar quantidade de valores
			if len(s.Values) != len(multiRet.Types) {
				c.reportError(0, 0, fmt.Sprintf("Function returns %d values, but return statement has %d",
					len(multiRet.Types), len(s.Values)))
				return
			}

			// Verificar cada valor individualmente
			for i, val := range s.Values {
				valType := c.checkExpr(val)
				expectedType := multiRet.Types[i]

				if !AreTypesCompatible(expectedType, valType) {
					c.reportError(0, 0, fmt.Sprintf("Type mismatch in return value %d. Expected %s, got %s",
						i+1, StringifyType(expectedType), StringifyType(valType)))
				}
			}
		} else {
			// Função retorna único valor
			if len(s.Values) > 1 {
				c.reportError(0, 0, "Function returns single value, but return statement has multiple values")
				return
			}

			valType := c.checkExpr(s.Values[0])
			if !AreTypesCompatible(c.currentFuncReturnType, valType) {
				c.reportError(0, 0, fmt.Sprintf("Type mismatch in return value. Expected %s, got %s",
					StringifyType(c.currentFuncReturnType), StringifyType(valType)))
			}
		}

	case *parser.BreakStmt:
		if !c.inLoop {
			c.reportError(0, 0, "'break' is only allowed inside loops")
		}

	case *parser.ContinueStmt:
		if !c.inLoop {
			c.reportError(0, 0, "'continue' is only allowed inside loops")
		}

	case *parser.ExprStmt:
		c.checkExpr(s.Expr)

	case *parser.PackageDecl:
		// Nada a verificar para declaração de pacote
		break

	case *parser.ImportDecl:
		c.checkImportDecl(s)

	case *parser.ExportDecl:
		c.checkExportDecl(s)
	}
}

func (c *Checker) checkConstDecl(decl *parser.ConstDecl) {
	initType := c.checkExpr(decl.Init)

	sym := &Symbol{
		Name: decl.Name,
		Kind: KindConst,
		Type: initType,
		Node: decl,
	}

	if !c.CurrentScope.Define(decl.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Constant '%s' redeclared in this scope", decl.Name))
	}
}

// check_smtp.go - Modificar a função checkVarDecl

func (c *Checker) checkVarDecl(decl *parser.VarDecl) {
	var initType Type
	if decl.Init != nil {
		initType = c.checkExpr(decl.Init)

		// Se o tipo do inicializador for "error", não prosseguir
		if StringifyType(initType) == "error" {
			return
		}

		if decl.Type == nil {
			// Inferência de tipo
			decl.Type = c.unwrapType(initType)
		} else {
			// Resolver o tipo declarado (pode ser um alias como "Number")
			resolvedDeclType := c.resolveType(decl.Type)
			declType := c.wrapType(resolvedDeclType)
			if !c.areTypesCompatible(declType, initType) {
				c.reportError(0, 0, fmt.Sprintf("Cannot assign type %s to variable '%s' of type %s",
					StringifyType(initType), decl.Name, StringifyType(declType)))
			}
		}
	}

	var symType Type
	if decl.Type != nil {
		// Para o símbolo, armazenamos o tipo declarado (não resolvido)
		// mas verificamos a compatibilidade com o tipo resolvido
		symType = c.wrapType(decl.Type)
	} else if initType != nil {
		symType = initType
	} else {
		// Variável sem tipo e sem inicializador (declaração forward)
		symType = &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}
	}

	sym := &Symbol{
		Name: decl.Name,
		Kind: KindVar,
		Type: symType,
		Node: decl,
	}

	if !c.CurrentScope.Define(decl.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Variable '%s' already declared in this scope", decl.Name))
	}
}

// checkMultiVarDecl processa declaração múltipla de variáveis
func (c *Checker) checkMultiVarDecl(decl *parser.MultiVarDecl) {
	// Verificar inicializador
	if decl.Init == nil {
		c.reportError(0, 0, "Multiple variables declaration must have initializer")
		return
	}

	initType := c.checkExpr(decl.Init)
	if initType == nil {
		c.reportError(0, 0, "Invalid initializer in multi-variable declaration")
		return
	}

	// Verificar se o inicializador retorna múltiplos valores
	var valueTypes []Type
	if multiType, ok := initType.(*MultiValueType); ok {
		valueTypes = multiType.Types
	} else {
		// Se não for MultiValueType, todos recebem o mesmo tipo
		valueTypes = make([]Type, len(decl.Names))
		for i := range valueTypes {
			valueTypes[i] = initType
		}
	}

	// Verificar compatibilidade de quantidade
	if len(decl.Names) != len(valueTypes) {
		c.reportError(0, 0, fmt.Sprintf("Mismatch in variable count: declared %d, but initializer provides %d values",
			len(decl.Names), len(valueTypes)))
		return
	}

	// Definir cada variável
	for i, name := range decl.Names {
		var symType Type
		if decl.Type != nil {
			symType = c.wrapType(decl.Type)
		} else {
			symType = valueTypes[i]
		}

		sym := &Symbol{
			Name: name,
			Kind: KindVar,
			Type: symType,
			Node: decl,
		}

		if !c.CurrentScope.Define(name, sym) {
			c.reportError(0, 0, fmt.Sprintf("Variable '%s' already declared in this scope", name))
		}
	}
}

// checkMultiConstDecl processa declaração múltipla de constantes
func (c *Checker) checkMultiConstDecl(decl *parser.MultiConstDecl) {
	// Similar ao checkMultiVarDecl, mas para constantes
	if decl.Init == nil {
		c.reportError(0, 0, "Multiple constants declaration must have initializer")
		return
	}

	initType := c.checkExpr(decl.Init)
	if initType == nil {
		c.reportError(0, 0, "Invalid initializer in multi-constant declaration")
		return
	}

	var valueTypes []Type
	if multiType, ok := initType.(*MultiValueType); ok {
		valueTypes = multiType.Types
	} else {
		valueTypes = make([]Type, len(decl.Names))
		for i := range valueTypes {
			valueTypes[i] = initType
		}
	}

	if len(decl.Names) != len(valueTypes) {
		c.reportError(0, 0, fmt.Sprintf("Mismatch in constant count: declared %d, but initializer provides %d values",
			len(decl.Names), len(valueTypes)))
		return
	}

	for i, name := range decl.Names {
		sym := &Symbol{
			Name: name,
			Kind: KindConst,
			Type: valueTypes[i],
			Node: decl,
		}

		if !c.CurrentScope.Define(name, sym) {
			c.reportError(0, 0, fmt.Sprintf("Constant '%s' already declared in this scope", name))
		}
	}
}

func (c *Checker) checkFunctionDecl(fn *parser.FunctionDecl) {
	// Converter tipos de retorno para nosso formato
	returnType := ToMultiValueType(fn.ReturnTypes)

	sym := &Symbol{Name: fn.Name, Kind: KindFunction, Type: returnType, Node: fn}
	if !c.CurrentScope.Define(fn.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Function '%s' redeclared", fn.Name))
	}

	c.enterScope()
	prevReturn := c.currentFuncReturnType
	c.currentFuncReturnType = returnType

	// Generics - registrar parâmetros genéricos como tipos
	if fn.Generics != nil {
		for _, g := range fn.Generics {
			// Registrar como um tipo genérico
			genericType := &ParserTypeWrapper{
				Type: &parser.GenericType{
					Name:     g.Name,
					TypeArgs: []parser.Type{},
				},
			}
			c.CurrentScope.Define(g.Name, &Symbol{
				Name: g.Name,
				Kind: KindGenericParam,
				Type: genericType,
			})
		}
	}

	// Params
	for _, param := range fn.Params {
		// Validar se o tipo do parâmetro existe
		c.validateTypeExists(param.Type)
		paramType := c.wrapType(param.Type)
		c.CurrentScope.Define(param.Name, &Symbol{Name: param.Name, Kind: KindVar, Type: paramType})
	}

	for _, stmt := range fn.Body {
		c.checkStmt(stmt)
	}

	c.currentFuncReturnType = prevReturn
	c.exitScope()
}

func (c *Checker) checkStructDecl(s *parser.StructDecl) {
	sym := &Symbol{
		Name: s.Name,
		Kind: KindStruct,
		Type: nil, // Structs não têm tipo semântico direto (é apenas uma definição)
		Node: s,
	}
	// Define o struct no escopo atual (geralmente global)
	if !c.CurrentScope.Define(s.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Struct '%s' already defined", s.Name))
	}

	// Cria um escopo temporário para validar os campos
	// Isso permite que parâmetros genéricos (T) sejam visíveis dentro do struct
	c.enterScope()

	// Registra genéricos no escopo temporário
	if s.Generics != nil {
		for _, g := range s.Generics {
			c.CurrentScope.Define(g.Name, &Symbol{Name: g.Name, Kind: KindGenericParam})
		}
	}

	// Validar campos
	fieldNames := make(map[string]bool)
	for _, field := range s.Fields {
		c.validateTypeExists(field.Type)
		if fieldNames[field.Name] {
			c.reportError(0, 0, fmt.Sprintf("Duplicate field '%s' in struct '%s'", field.Name, s.Name))
		}
		fieldNames[field.Name] = true
	}

	c.exitScope()
}

func (c *Checker) checkTypeDecl(s *parser.TypeDecl) {
	c.validateTypeExists(s.Type)

	// Resolver o tipo para garantir que aliases aninhados sejam expandidos
	resolvedType := c.resolveType(s.Type)

	typeSym := &Symbol{
		Name: s.Name,
		Kind: KindTypeAlias,
		Type: c.wrapType(resolvedType), // Armazenar o tipo resolvido
		Node: s,
	}

	if !c.CurrentScope.Define(s.Name, typeSym) {
		c.reportError(0, 0, fmt.Sprintf("Type '%s' already defined", s.Name))
	}
}

func (c *Checker) checkImplDecl(s *parser.ImplDecl) {
	// Verificar se o Target existe
	sym := c.CurrentScope.Resolve(s.TargetName)
	if sym == nil || sym.Kind != KindStruct {
		c.reportError(0, 0, fmt.Sprintf("Cannot implement methods for unknown struct '%s'", s.TargetName))
		return
	}

	// Simplificação: Checar os métodos como funções normais,
	// mas injetando 'self' no escopo se necessário.
	for _, method := range s.Methods {
		c.checkMethodDecl(method, sym.Type)
	}
}

func (c *Checker) checkMethodDecl(m *parser.MethodDecl, structType Type) {
	c.enterScope()

	// Definir 'self'
	c.CurrentScope.Define("self", &Symbol{Name: "self", Kind: KindVar, Type: structType})

	prevReturn := c.currentFuncReturnType
	c.currentFuncReturnType = ToMultiValueType(m.ReturnTypes)

	// Generics do método
	if m.Generics != nil {
		for _, g := range m.Generics {
			c.CurrentScope.Define(g.Name, &Symbol{Name: g.Name, Kind: KindGenericParam})
		}
	}

	// Params
	for _, param := range m.Params {
		c.validateTypeExists(param.Type)
		paramType := c.wrapType(param.Type)
		c.CurrentScope.Define(param.Name, &Symbol{Name: param.Name, Kind: KindVar, Type: paramType})
	}

	for _, stmt := range m.Body {
		c.checkStmt(stmt)
	}

	c.currentFuncReturnType = prevReturn
	c.exitScope()
}

func (c *Checker) checkSwitchStmt(s *parser.SwitchStmt) {
	exprType := c.checkExpr(s.Expr)

	for _, clause := range s.Cases {
		if clause.Value != nil {
			caseType := c.checkExpr(clause.Value)
			if !c.areTypesCompatible(exprType, caseType) {
				c.reportError(0, 0, fmt.Sprintf("Case type mismatch. Switch on %s, but case is %s",
					StringifyType(exprType), StringifyType(caseType)))
			}
		}
		c.checkBlockScope(clause.Body)
	}
}

// Helpers
func (c *Checker) checkBlockScope(stmts []parser.Stmt) {
	c.enterScope()
	for _, s := range stmts {
		c.checkStmt(s)
	}
	c.exitScope()
}

func (c *Checker) checkBlock(stmts []parser.Stmt) {
	c.checkBlockScope(stmts)
}

func (c *Checker) checkImportDecl(imp *parser.ImportDecl) {
	if imp.Imports != nil {
		for _, spec := range imp.Imports {
			symbolName := spec.Name
			if spec.Alias != "" {
				symbolName = spec.Alias
			}
			sym := &Symbol{Name: symbolName, Kind: KindImport, Node: imp}
			if !c.CurrentScope.Define(symbolName, sym) {
				c.reportError(0, 0, fmt.Sprintf("Import '%s' already declared", symbolName))
			}
		}
	} else {
		path := imp.Path
		parts := strings.Split(path, ".")
		moduleName := parts[len(parts)-1]

		sym := &Symbol{
			Name: moduleName,
			Kind: KindImport,
			Node: imp,
		}

		if !c.CurrentScope.Define(moduleName, sym) {
			c.reportError(0, 0, fmt.Sprintf("Module '%s' already declared", moduleName))
		}
	}
}

func (c *Checker) checkExportDecl(exp *parser.ExportDecl) {
	for _, spec := range exp.Exports {
		sym := c.CurrentScope.Resolve(spec.Name)
		if sym == nil {
			c.reportError(0, 0, fmt.Sprintf("Cannot export undeclared symbol '%s'", spec.Name))
		}
	}
}

func (c *Checker) validateTypeExists(t parser.Type) {
	switch v := t.(type) {
	case *parser.IdentifierType:
		if c.CurrentScope.Resolve(v.Name) == nil {
			c.reportError(0, 0, fmt.Sprintf("Unknown type '%s'", v.Name))
		}
	case *parser.ArrayType:
		c.validateTypeExists(v.ElementType)
	case *parser.PrimitiveType:
		// Tipos primitivos sempre existem
		return
	case *parser.NullableType:
		c.validateTypeExists(v.BaseType)
	case *parser.MapType:
		c.validateTypeExists(v.KeyType)
		c.validateTypeExists(v.ValueType)
	case *parser.PointerType:
		c.validateTypeExists(v.BaseType)
	case *parser.UnionType:
		for _, typ := range v.Types {
			c.validateTypeExists(typ)
		}
	}
}

// Helper functions
func (c *Checker) isBooleanType(t Type) bool {
	typeStr := StringifyType(t)
	return typeStr == "bool" || typeStr == "bool?"
}

func (c *Checker) isConditionableType(t Type) bool {
	if t == nil {
		return false
	}

	typeStr := StringifyType(t)

	// Tipos booleanos (com ou sem nullable)
	if typeStr == "bool" || typeStr == "bool?" {
		return true
	}

	// Tipos numéricos (com ou sem nullable)
	if typeStr == "int" || typeStr == "int?" ||
		typeStr == "float" || typeStr == "float?" {
		return true
	}

	// Outros tipos nullable (qualquer tipo com ?)
	if strings.HasSuffix(typeStr, "?") {
		return true
	}

	return false
}

func (c *Checker) areTypesCompatible(target, source Type) bool {
	// Implementação simplificada - deve usar AreParserTypesCompatible para tipos wrapped
	if target == nil || source == nil {
		return false
	}

	// Ambos são wrappers de parser.Type
	if wTarget, ok := target.(*ParserTypeWrapper); ok {
		if wSource, ok := source.(*ParserTypeWrapper); ok {
			return AreParserTypesCompatible(wTarget.Type, wSource.Type)
		}
	}

	// MultiValueType
	if mTarget, ok := target.(*MultiValueType); ok {
		if mSource, ok := source.(*MultiValueType); ok {
			if len(mTarget.Types) != len(mSource.Types) {
				return false
			}
			for i := range mTarget.Types {
				if !c.areTypesCompatible(mTarget.Types[i], mSource.Types[i]) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Tipos diferentes
	return false
}

func (c *Checker) unwrapType(t Type) parser.Type {
	if w, ok := t.(*ParserTypeWrapper); ok {
		return w.Type
	}
	// Não é um tipo wrapped, retornar nil
	return nil
}
