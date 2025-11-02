package semantic

import (
	"fmt"

	"github.com/alpha/internal/parser"
)

// Checker principal que percorre AST e coleta erros
type Checker struct {
	global *Scope
	errors []SemanticErr

	// opcional: map de nós para tipos resolvidos (útil para codegen)
	Types map[parser.Node]Type
}

// NewChecker cria uma instância vazia
func NewChecker() *Checker {
	return &Checker{
		global: NewScope(nil),
		errors: []SemanticErr{},
		Types:  make(map[parser.Node]Type),
	}
}

// addError registra um erro semântico
func (c *Checker) addError(msg string, line, col int) {
	c.errors = append(c.errors, SemanticErr{Msg: msg, Line: line, Col: col})
}

// Check é a API pública: recebe AST do parser e retorna erros encontrados
func (c *Checker) Check(prog *parser.Program) []SemanticErr {
	// reset
	c.global = NewScope(nil)
	c.errors = []SemanticErr{}
	c.Types = make(map[parser.Node]Type)

	// declarar builtins mínimos
	_ = c.global.Define(&Symbol{
		Name:     "print",
		Type:     FuncType{Params: []Type{AnyType{}}, Ret: AnyType{}},
		Mutable:  false,
		DeclLine: 0,
		DeclCol:  0,
	})

	// percorrer o corpo do programa
	for _, stmt := range prog.Body {
		c.checkStmt(stmt, c.global)
	}
	return c.errors
}

// checkStmt faz a análise semântica de uma statement dentro do escopo sc
func (c *Checker) checkStmt(s parser.Stmt, sc *Scope) {
	switch st := s.(type) {
	case *parser.VarDecl:
		// inferir tipo do init, se presente; senão Any
		var typ Type = AnyType{}
		if st.Init != nil {
			typ = c.checkExpr(st.Init, sc)
		}
		sym := &Symbol{
			Name:     st.Name,
			Type:     typ,
			Mutable:  true,
			DeclLine: 0,
			DeclCol:  0,
		}
		if err := sc.Define(sym); err != nil {
			if serr, ok := err.(*SemanticErr); ok {
				c.addError(serr.Msg, serr.Line, serr.Col)
			} else {
				c.addError(err.Error(), 0, 0)
			}
		}
	case *parser.ConstDecl:
		if st.Init == nil {
			c.addError("const declaration must have an initializer", 0, 0)
			return
		}
		typ := c.checkExpr(st.Init, sc)
		sym := &Symbol{
			Name:     st.Name,
			Type:     typ,
			Mutable:  false,
			DeclLine: 0,
			DeclCol:  0,
		}
		if err := sc.Define(sym); err != nil {
			if serr, ok := err.(*SemanticErr); ok {
				c.addError(serr.Msg, serr.Line, serr.Col)
			} else {
				c.addError(err.Error(), 0, 0)
			}
		}
	case *parser.ExprStmt:
		c.checkExpr(st.Expr, sc)
	case *parser.IfStmt:
		condType := c.checkExpr(st.Cond, sc)
		if _, ok := condType.(BoolType); !ok {
			c.addError("if condition must be boolean", 0, 0)
		}
		// then block with new inner scope
		thenScope := NewScope(sc)
		for _, s2 := range st.Then {
			c.checkStmt(s2, thenScope)
		}
		// else block
		if st.Else != nil {
			elseScope := NewScope(sc)
			for _, s2 := range st.Else {
				c.checkStmt(s2, elseScope)
			}
		}
	case *parser.WhileStmt:
		condType := c.checkExpr(st.Cond, sc)
		if _, ok := condType.(BoolType); !ok {
			c.addError("while condition must be boolean", 0, 0)
		}
		bodyScope := NewScope(sc)
		for _, s2 := range st.Body {
			c.checkStmt(s2, bodyScope)
		}
	case *parser.ForStmt:
		// for init; cond; post { body }
		loopScope := NewScope(sc)
		if st.Init != nil {
			// init can be Stmt (VarDecl or ExprStmt)
			c.checkStmt(st.Init, loopScope)
		}
		if st.Cond != nil {
			ct := c.checkExpr(st.Cond, loopScope)
			if _, ok := ct.(BoolType); !ok {
				c.addError("for condition must be boolean", 0, 0)
			}
		}
		if st.Post != nil {
			c.checkStmt(st.Post, loopScope)
		}
		for _, s2 := range st.Body {
			c.checkStmt(s2, loopScope)
		}
	case *parser.ReturnStmt:
		// sem suporte a funções completas por enquanto; apenas checa expressão
		if st.Value != nil {
			c.checkExpr(st.Value, sc)
		}
	case *parser.BlockStmt:
		blockScope := NewScope(sc)
		for _, s2 := range st.Body {
			c.checkStmt(s2, blockScope)
		}
	default:
		// declaração desconhecida: ignorar para robustez
	}
}

// checkExpr resolve tipos de expressões e faz validações.
// Retorna o tipo inferido (ou AnyType como fallback).
func (c *Checker) checkExpr(e parser.Expr, sc *Scope) Type {
	// registrar tipo em map opcional: c.Types[node] = type
	switch ex := e.(type) {
	case *parser.IntLiteral:
		// opcional: associar nó ao tipo
		return IntType{}
	case *parser.FloatLiteral:
		return FloatType{}
	case *parser.StringLiteral:
		return StringType{}
	case *parser.BoolLiteral:
		return BoolType{}
	case *parser.NullLiteral:
		return NullType{}
	case *parser.Identifier:
		if sym, ok := sc.Resolve(ex.Name); ok {
			return sym.Type
		}
		c.addError(fmt.Sprintf("undefined identifier %q", ex.Name), 0, 0)
		return AnyType{}
	case *parser.BinaryExpr:
		lt := c.checkExpr(ex.Left, sc)
		rt := c.checkExpr(ex.Right, sc)
		switch ex.Op {
		case "+", "-", "*", "/":
			// números: int/float coercion
			if _, li := lt.(IntType); li {
				if _, ri := rt.(IntType); ri {
					return IntType{}
				}
				if _, rf := rt.(FloatType); rf {
					return FloatType{}
				}
			}
			if _, lf := lt.(FloatType); lf {
				if _, rf := rt.(FloatType); rf {
					return FloatType{}
				}
				if _, ri := rt.(IntType); ri {
					return FloatType{}
				}
			}
			// string + string
			if ex.Op == "+" {
				if _, ls := lt.(StringType); ls {
					if _, rs := rt.(StringType); rs {
						return StringType{}
					}
				}
			}
			c.addError("invalid operand types for arithmetic operator "+ex.Op, 0, 0)
			return AnyType{}
		case "==", "!=", "<", "<=", ">", ">=":
			// comparisons produce bool; allow numeric comparisons and string equality
			if (lt.Equals(IntType{}) || lt.Equals(FloatType{}) || lt.Equals(StringType{})) &&
				(rt.Equals(IntType{}) || rt.Equals(FloatType{}) || rt.Equals(StringType{})) {
				return BoolType{}
			}
			c.addError("invalid operand types for comparison "+ex.Op, 0, 0)
			return AnyType{}
		case "&&", "||":
			if _, lBool := lt.(BoolType); lBool {
				if _, rBool := rt.(BoolType); rBool {
					return BoolType{}
				}
			}
			c.addError("logical operators require boolean operands", 0, 0)
			return AnyType{}
		default:
			c.addError("unknown binary operator "+ex.Op, 0, 0)
			return AnyType{}
		}
	case *parser.AssignExpr:
		// left deve ser um identifier (simples) para esse checker básico
		if id, ok := ex.Left.(*parser.Identifier); ok {
			sym, found := sc.Resolve(id.Name)
			if !found {
				c.addError(fmt.Sprintf("assign to undefined variable %q", id.Name), 0, 0)
				return AnyType{}
			}
			if !sym.Mutable {
				c.addError(fmt.Sprintf("cannot assign to constant %q", id.Name), 0, 0)
				return AnyType{}
			}
			rt := c.checkExpr(ex.Right, sc)
			// simples verificação de compatibilidade (aceita promoção int->float)
			if sym.Type.Equals(rt) {
				return sym.Type
			}
			// allow assign int -> float
			if _, ok1 := sym.Type.(FloatType); ok1 {
				if _, ok2 := rt.(IntType); ok2 {
					return sym.Type
				}
			}
			c.addError("type mismatch in assignment", 0, 0)
			return AnyType{}
		}
		c.addError("left-hand side of assignment must be an identifier", 0, 0)
		return AnyType{}
	case *parser.CallExpr:
		// checar o tipo do callee
		calleeType := c.checkExpr(ex.Callee, sc)
		ft, ok := calleeType.(FuncType)
		if !ok {
			c.addError("call of non-function value", 0, 0)
			// ainda percorre args para acumular possíveis erros
			for _, a := range ex.Args {
				c.checkExpr(a, sc)
			}
			return AnyType{}
		}
		// checar aridade
		if len(ex.Args) != len(ft.Params) {
			c.addError("wrong number of arguments in call", 0, 0)
			for _, a := range ex.Args {
				c.checkExpr(a, sc)
			}
			return ft.Ret
		}
		// checar tipos dos argumentos
		for i, a := range ex.Args {
			at := c.checkExpr(a, sc)
			if !ft.Params[i].Equals(at) && !(ft.Params[i].Equals(FloatType{}) && at.Equals(IntType{})) {
				c.addError("argument type mismatch", 0, 0)
			}
		}
		return ft.Ret
	default:
		// casos não previstos -> any
		return AnyType{}
	}
}
