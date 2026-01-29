package codegen

import (
	"fmt"
	"strings"

	"github.com/alpha/internal/parser"
	"github.com/alpha/internal/semantic"
)

// TypeMapper gerencia conversões de tipos Alpha -> Go
type TypeMapper struct {
	structTypes map[string]string // cache de tipos de struct
	unionTypes  map[string]string // cache de union types
}

func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		structTypes: make(map[string]string),
		unionTypes:  make(map[string]string),
	}
}

// ToGoType converte tipos Alpha para Go mantendo Generics
func (tm *TypeMapper) ToGoType(t semantic.Type) string {
	if t == nil {
		return "" // void
	}

	// Desembrulha o ParserTypeWrapper se necessário
	if wrapper, ok := t.(*semantic.ParserTypeWrapper); ok {
		return tm.mapParserType(wrapper.Type)
	}

	// Fallback para tipos manuais ou strings
	return tm.mapSimpleString(semantic.StringifyType(t))
}

func (tm *TypeMapper) mapSimpleString(s string) string {
	switch s {
	case "int":
		return "int64"
	case "float":
		return "float64"
	case "bool":
		return "bool"
	case "string":
		return "string"
	case "void":
		return ""
	case "any":
		return "interface{}"
	case "byte":
		return "byte"
	case "char":
		return "rune"
	case "error":
		return "error"
	default:
		// Presume ser um tipo definido pelo usuário
		return s
	}
}

// GetElementType extrai o tipo elemento de arrays/slices
func (tm *TypeMapper) GetElementType(goType string) string {
	if strings.HasPrefix(goType, "[]") {
		return goType[2:]
	}
	if strings.HasPrefix(goType, "map[") {
		// Extrai tipo valor de map[K]V
		parts := strings.SplitN(goType, "]", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return goType
}

// IsReferenceType verifica se o tipo é passado por referência
func (tm *TypeMapper) IsReferenceType(goType string) bool {
	return strings.HasPrefix(goType, "*") ||
		strings.HasPrefix(goType, "[]") ||
		strings.HasPrefix(goType, "map[") ||
		strings.HasPrefix(goType, "interface{}")
}

// ZeroValue retorna valor zero otimizado
func (tm *TypeMapper) ZeroValue(goType string) string {
	switch goType {
	case "int64", "float64":
		return "0"
	case "bool":
		return "false"
	case "string":
		return `""`
	case "byte":
		return "0"
	case "rune":
		return "0"
	default:
		if strings.HasPrefix(goType, "*") ||
			strings.HasPrefix(goType, "[]") ||
			strings.HasPrefix(goType, "map[") ||
			goType == "interface{}" ||
			goType == "error" {
			return "nil"
		}
		// Para structs, retorna struct literal vazia
		return goType + "{}"
	}
}

// CanUseStack verifica se o tipo pode ser alocado na stack
func (tm *TypeMapper) CanUseStack(goType string) bool {
	// Tipos pequenos podem ir para stack
	smallTypes := map[string]bool{
		"int64":   true,
		"float64": true,
		"bool":    true,
		"byte":    true,
		"rune":    true,
	}

	if smallTypes[goType] {
		return true
	}

	// Ponteiros para tipos pequenos também
	if strings.HasPrefix(goType, "*") {
		base := goType[1:]
		return smallTypes[base] || base == "string"
	}

	return false
}

func (tm *TypeMapper) mapParserType(t parser.Type) string {
	switch pt := t.(type) {
	case *parser.PrimitiveType:
		switch pt.Name {
		case "int":
			return "int"
		case "float":
			return "float64"
		case "bool":
			return "bool"
		case "string":
			return "string"
		case "byte":
			return "byte"
		case "char":
			return "rune"
		case "error":
			return "error"
		case "void":
			return ""
		default:
			return "interface{}"
		}

	case *parser.IdentifierType:
		// Tipos definidos pelo usuário
		return pt.Name

	case *parser.GenericType:
		// Para tipos genéricos, usamos interface{} como fallback
		if len(pt.TypeArgs) == 0 {
			return pt.Name
		}
		return "interface{}"

	case *parser.ArrayType:
		elemType := tm.mapParserType(pt.ElementType)
		return "[]" + elemType

	case *parser.MapType:
		keyType := tm.mapParserType(pt.KeyType)
		valueType := tm.mapParserType(pt.ValueType)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)

	case *parser.PointerType:
		baseType := tm.mapParserType(pt.BaseType)
		return "*" + baseType

	case *parser.NullableType:
		baseType := tm.mapParserType(pt.BaseType)
		return "*" + baseType

	case *parser.UnionType:
		// Para union types, usamos interface{}
		return "interface{}"

	case *parser.SetType:
		elemType := tm.mapParserType(pt.ElementType)
		return fmt.Sprintf("map[%s]struct{}", elemType)

	default:
		return "interface{}"
	}
}
