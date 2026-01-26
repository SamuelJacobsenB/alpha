package semantic

import (
	"github.com/alpha/internal/parser"
)

type Checker struct {
	CurrentScope *Scope
	Errors       []SemanticError

	// Contexto atual
	currentFuncReturnType Type
	inLoop                bool
}

// checker.go - função NewChecker()
func NewChecker() *Checker {
	global := NewScope(nil)

	// Registrar tipos primitivos básicos como aliases
	global.Define("int", &Symbol{Name: "int", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "int"}}})
	global.Define("float", &Symbol{Name: "float", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "float"}}})
	global.Define("string", &Symbol{Name: "string", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "string"}}})
	global.Define("bool", &Symbol{Name: "bool", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "bool"}}})
	global.Define("void", &Symbol{Name: "void", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})
	global.Define("any", &Symbol{Name: "any", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "any"}}})

	// Adicionar tipos nullable básicos
	global.Define("int?", &Symbol{Name: "int?", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.NullableType{BaseType: &parser.PrimitiveType{Name: "int"}}}})
	global.Define("float?", &Symbol{Name: "float?", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.NullableType{BaseType: &parser.PrimitiveType{Name: "float"}}}})
	global.Define("bool?", &Symbol{Name: "bool?", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.NullableType{BaseType: &parser.PrimitiveType{Name: "bool"}}}})
	global.Define("string?", &Symbol{Name: "string?", Kind: KindTypeAlias, Type: &ParserTypeWrapper{Type: &parser.NullableType{BaseType: &parser.PrimitiveType{Name: "string"}}}})

	// REGISTRAR FUNÇÕES BUILT-IN
	// Funções de array
	global.Define("append", &Symbol{Name: "append", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})
	global.Define("remove", &Symbol{Name: "remove", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})
	global.Define("removeIndex", &Symbol{Name: "removeIndex", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})
	global.Define("length", &Symbol{Name: "length", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "int"}}})

	// Funções de mapa
	global.Define("delete", &Symbol{Name: "delete", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})
	global.Define("clear", &Symbol{Name: "clear", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})

	// Funções de set
	global.Define("has", &Symbol{Name: "has", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "bool"}}})
	global.Define("add", &Symbol{Name: "add", Kind: KindFunction, Type: &ParserTypeWrapper{Type: &parser.PrimitiveType{Name: "void"}}})

	return &Checker{
		CurrentScope: global,
		Errors:       make([]SemanticError, 0),
		inLoop:       false,
	}
}

func (c *Checker) CheckProgram(prog *parser.Program) {
	for _, stmt := range prog.Body {
		c.checkStmt(stmt)
	}
}

func (c *Checker) reportError(line, col int, msg string) {
	c.Errors = append(c.Errors, SemanticError{
		Msg:  msg,
		Line: line,
		Col:  col,
	})
}

func (c *Checker) enterScope() {
	c.CurrentScope = NewScope(c.CurrentScope)
}

func (c *Checker) exitScope() {
	if c.CurrentScope.Outer != nil {
		c.CurrentScope = c.CurrentScope.Outer
	}
}

// Helper para converter parser.Type para semantic.Type
func (c *Checker) wrapType(t parser.Type) Type {
	return ToType(t)
}

func (c *Checker) resolveType(t parser.Type) parser.Type {
	if t == nil {
		return nil
	}

	switch v := t.(type) {
	case *parser.IdentifierType:
		sym := c.CurrentScope.Resolve(v.Name)
		if sym != nil && sym.Kind == KindTypeAlias {
			// Se for um alias, retorna o tipo subjacente
			if wrapper, ok := sym.Type.(*ParserTypeWrapper); ok {
				return wrapper.Type
			}
		}
		// Não é um alias ou não foi encontrado
		return t

	case *parser.ArrayType:
		return &parser.ArrayType{ElementType: c.resolveType(v.ElementType)}

	case *parser.PointerType:
		return &parser.PointerType{BaseType: c.resolveType(v.BaseType)}

	case *parser.NullableType:
		return &parser.NullableType{BaseType: c.resolveType(v.BaseType)}

	case *parser.MapType:
		return &parser.MapType{
			KeyType:   c.resolveType(v.KeyType),
			ValueType: c.resolveType(v.ValueType),
		}

	case *parser.UnionType:
		types := make([]parser.Type, len(v.Types))
		for i, typ := range v.Types {
			types[i] = c.resolveType(typ)
		}
		return &parser.UnionType{Types: types}

	case *parser.SetType:
		return &parser.SetType{ElementType: c.resolveType(v.ElementType)}

	case *parser.GenericType:
		args := make([]parser.Type, len(v.TypeArgs))
		for i, arg := range v.TypeArgs {
			args[i] = c.resolveType(arg)
		}
		return &parser.GenericType{Name: v.Name, TypeArgs: args}

	default:
		// Para PrimitiveType e outros, retorna como está
		return t
	}
}
