package semantic

import (
	"fmt"

	"github.com/alpha/internal/parser"
)

// Helper para converter tipos primitivos em string para comparação rápida
func getTypeName(t parser.Type) string {
	if t == nil {
		return "void"
	} // Proteção contra nil
	switch v := t.(type) {
	case *parser.PrimitiveType:
		return v.Name
	case *parser.IdentifierType:
		return v.Name
	case *parser.ArrayType:
		return getTypeName(v.ElementType) + "[]"
	case *parser.NullableType:
		return getTypeName(v.BaseType) + "?"
	default:
		return "unknown"
	}
}

// StringifyType converte qualquer nó de tipo em string legível
func StringifyType(t parser.Type) string {
	if t == nil {
		return "void"
	}
	switch v := t.(type) {
	case *parser.PrimitiveType:
		return v.Name
	case *parser.IdentifierType:
		return v.Name
	case *parser.ArrayType:
		return StringifyType(v.ElementType) + "[]"
	case *parser.PointerType:
		return "*" + StringifyType(v.BaseType)
	case *parser.NullableType:
		return StringifyType(v.BaseType) + "?"
	case *parser.MapType:
		return fmt.Sprintf("map<%s, %s>", StringifyType(v.KeyType), StringifyType(v.ValueType))
	default:
		return "unknown"
	}
}

func isConditionable(t parser.Type) bool {
	tName := StringifyType(t)
	// Aceita bool, int, float
	if tName == "bool" || tName == "int" || tName == "float" {
		return true
	}
	// Aceita Nullable (verifica se não é null)
	if _, ok := t.(*parser.NullableType); ok {
		return true
	}
	return false
}

// isNumeric: Helper reutilizado
func isNumeric(t parser.Type) bool {
	tBase := unwrapNullable(t)
	name := StringifyType(tBase)
	return name == "int" || name == "float"
}

// unwrapNullable: Remove o wrapper nullable se existir
func unwrapNullable(t parser.Type) parser.Type {
	if nt, ok := t.(*parser.NullableType); ok {
		return nt.BaseType
	}
	return t
}

// AreTypesCompatible: Mantido igual, garantindo coerção de nullable e numérico
func AreTypesCompatible(target, source parser.Type) bool {
	if target == nil || source == nil {
		return false
	}

	tName := StringifyType(target)
	sName := StringifyType(source)

	if tName == "error" || sName == "error" {
		return true
	}
	if tName == sName {
		return true
	}

	tBase := unwrapNullable(target)
	sBase := unwrapNullable(source)

	if StringifyType(tBase) == StringifyType(sBase) {
		return true
	}

	// Permite int -> float
	if StringifyType(tBase) == "float" && StringifyType(sBase) == "int" {
		return true
	}

	return false
}
