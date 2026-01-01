package semantic

import (
	"fmt"

	"github.com/alpha/internal/parser"
)

func (c *Checker) checkStmt(stmt parser.Stmt) {
	switch s := stmt.(type) {

	case *parser.ConstDecl:
		c.checkConstDecl(s)

	case *parser.VarDecl:
		c.checkVarDecl(s)

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
		if !isConditionable(condType) {
			c.reportError(0, 0, fmt.Sprintf("Condition in 'if' must be boolean or nullable, got %s", StringifyType(condType)))
		}
		c.checkBlockScope(s.Then)
		if s.Else != nil {
			c.checkBlockScope(s.Else)
		}

	case *parser.WhileStmt:
		condType := c.checkExpr(s.Cond)
		if !isConditionable(condType) {
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
		if !isBoolean(condType) {
			c.reportError(0, 0, "Condition in 'do-while' must be boolean")
		}

	case *parser.ForStmt:
		c.enterScope() // For cria escopo para o Init
		if s.Init != nil {
			c.checkStmt(s.Init)
		}
		if s.Cond != nil {
			condType := c.checkExpr(s.Cond)
			if !isBoolean(condType) {
				c.reportError(0, 0, "Condition in 'for' must be boolean")
			}
		}
		if s.Post != nil {
			c.checkStmt(s.Post)
		}

		prevLoop := c.inLoop
		c.inLoop = true
		// Nota: s.Body já é []Stmt, mas não queremos criar OUTRO escopo além do que já criamos
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
		if s.Value != nil {
			valType := c.checkExpr(s.Value)
			if !AreTypesCompatible(c.currentFuncReturnType, valType) {
				c.reportError(0, 0, fmt.Sprintf("Type mismatch in return value. Expected %s, got %s",
					StringifyType(c.currentFuncReturnType), StringifyType(valType)))
			}
		} else {
			// Retorno vazio: verificar se função é void
			if StringifyType(c.currentFuncReturnType) != "void" {
				c.reportError(0, 0, "Non-void function must return a value")
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
	}
}

// --- IMPLEMENTAÇÕES NOVAS E CORRIGIDAS ---

func (c *Checker) checkConstDecl(decl *parser.ConstDecl) {
	// Constantes OBRIGATORIAMENTE têm inicializador no parser
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

func (c *Checker) checkVarDecl(decl *parser.VarDecl) {
	if decl.Init != nil {
		initType := c.checkExpr(decl.Init)
		if decl.Type == nil {
			decl.Type = initType // Inferência
		} else {
			if !AreTypesCompatible(decl.Type, initType) {
				c.reportError(0, 0, fmt.Sprintf("Cannot assign type %s to variable '%s' of type %s",
					getTypeName(initType), decl.Name, getTypeName(decl.Type)))
			}
		}
	}

	sym := &Symbol{
		Name: decl.Name,
		Kind: KindVar,
		Type: decl.Type,
		Node: decl,
	}

	if !c.CurrentScope.Define(decl.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Variable '%s' already declared in this scope", decl.Name))
	}
}

func (c *Checker) checkFunctionDecl(fn *parser.FunctionDecl) {
	sym := &Symbol{Name: fn.Name, Kind: KindFunction, Type: fn.ReturnType, Node: fn}
	if !c.CurrentScope.Define(fn.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Function '%s' redeclared", fn.Name))
	}

	c.enterScope()
	prevReturn := c.currentFuncReturnType
	c.currentFuncReturnType = fn.ReturnType

	// Generics
	if fn.Generics != nil {
		for _, g := range fn.Generics {
			c.CurrentScope.Define(g.Name, &Symbol{Name: g.Name, Kind: KindGenericParam})
		}
	}

	// Params
	for _, param := range fn.Params {
		// Validar se o tipo do parâmetro existe (ex: User u, struct User existe?)
		c.validateTypeExists(param.Type)
		c.CurrentScope.Define(param.Name, &Symbol{Name: param.Name, Kind: KindVar, Type: param.Type})
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
		Node: s,
	}
	if !c.CurrentScope.Define(s.Name, sym) {
		c.reportError(0, 0, fmt.Sprintf("Struct '%s' already defined", s.Name))
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
}

func (c *Checker) checkTypeDecl(s *parser.TypeDecl) {
	c.validateTypeExists(s.Type)
	sym := &Symbol{Name: s.Name, Kind: KindTypeAlias, Type: s.Type, Node: s}
	if !c.CurrentScope.Define(s.Name, sym) {
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
		c.checkMethodDecl(method, sym.Type) // Passamos o tipo do struct para resolver 'self'
	}
}

func (c *Checker) checkMethodDecl(m *parser.MethodDecl, structType parser.Type) {
	// Implementar lógica similar a functionDecl, mas adicionar 'self' ao escopo
	c.enterScope()

	// Definir 'self'
	c.CurrentScope.Define("self", &Symbol{Name: "self", Kind: KindVar, Type: structType})

	prevReturn := c.currentFuncReturnType
	c.currentFuncReturnType = m.ReturnType

	for _, param := range m.Params {
		c.validateTypeExists(param.Type)
		c.CurrentScope.Define(param.Name, &Symbol{Name: param.Name, Kind: KindVar, Type: param.Type})
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
			if !AreTypesCompatible(exprType, caseType) {
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
	// BlockStmt já espera criar escopo
	c.checkBlockScope(stmts)
}

func (c *Checker) validateTypeExists(t parser.Type) {
	// Verificar se Identifiers usados como tipo (ex: User u) existem na tabela
	switch v := t.(type) {
	case *parser.IdentifierType:
		if c.CurrentScope.Resolve(v.Name) == nil {
			c.reportError(0, 0, fmt.Sprintf("Unknown type '%s'", v.Name))
		}
	case *parser.ArrayType:
		c.validateTypeExists(v.ElementType)
	}
}

func isBoolean(t parser.Type) bool {
	return StringifyType(t) == "bool"
}
