package semantic

import (
	"fmt"
	"strings"

	"github.com/alpha/internal/parser"
)

// =================================================================
// TIPO LOCAL PARA REPRESENTAR TIPOS (NÃO IMPLEMENTA parser.Type)
// =================================================================

// Type é nossa própria interface de tipo que pode representar tanto
// tipos do parser quanto nossos tipos especiais
type Type interface {
	String() string
	// Podemos adicionar outros métodos necessários
}

// =================================================================
// TIPOS E CONSTANTES
// =================================================================

// MultiValueType representa múltiplos retornos (ex: string, string)
type MultiValueType struct {
	Types []Type
}

func (m *MultiValueType) String() string {
	if len(m.Types) == 0 {
		return "void"
	}
	parts := make([]string, len(m.Types))
	for i, sub := range m.Types {
		parts[i] = sub.String()
	}
	return strings.Join(parts, ", ")
}

// =================================================================
// WRAPPERS para tipos do parser
// =================================================================

// ParserTypeWrapper envolve um parser.Type para implementar nossa interface Type
type ParserTypeWrapper struct {
	Type parser.Type
}

func (w *ParserTypeWrapper) String() string {
	return StringifyParserType(w.Type)
}

// =================================================================
// FUNÇÕES AUXILIARES
// =================================================================

// ToType converte um parser.Type em nosso Type
func ToType(t parser.Type) Type {
	if t == nil {
		return &ParserTypeWrapper{Type: t}
	}
	return &ParserTypeWrapper{Type: t}
}

// ToMultiValueType converte uma lista de parser.Type em MultiValueType
func ToMultiValueType(types []parser.Type) Type {
	if len(types) == 1 {
		return ToType(types[0])
	}

	wrappedTypes := make([]Type, len(types))
	for i, t := range types {
		wrappedTypes[i] = ToType(t)
	}
	return &MultiValueType{Types: wrappedTypes}
}

// StringifyParserType converte um parser.Type em string
func StringifyParserType(t parser.Type) string {
	if t == nil {
		return "void"
	}

	switch v := t.(type) {
	case *parser.PrimitiveType:
		return v.Name
	case *parser.IdentifierType:
		return v.Name
	case *parser.ArrayType:
		return StringifyParserType(v.ElementType) + "[]"
	case *parser.PointerType:
		return "*" + StringifyParserType(v.BaseType)
	case *parser.NullableType:
		return StringifyParserType(v.BaseType) + "?"
	case *parser.MapType:
		return fmt.Sprintf("map<%s, %s>",
			StringifyParserType(v.KeyType),
			StringifyParserType(v.ValueType))
	case *parser.UnionType:
		typeStrings := make([]string, len(v.Types))
		for i, typ := range v.Types {
			typeStrings[i] = StringifyParserType(typ)
		}
		return strings.Join(typeStrings, " | ")
	case *parser.SetType:
		return fmt.Sprintf("set<%s>", StringifyParserType(v.ElementType))
	case *parser.GenericType:
		if len(v.TypeArgs) == 0 {
			return v.Name
		}
		args := make([]string, len(v.TypeArgs))
		for i, arg := range v.TypeArgs {
			args[i] = StringifyParserType(arg)
		}
		return fmt.Sprintf("%s<%s>", v.Name, strings.Join(args, ", "))
	default:
		return "unknown"
	}
}

// =================================================================
// FUNÇÕES DE COMPATIBILIDADE
// =================================================================

func AreTypesCompatible(target, source Type) bool {
	if target == nil || source == nil {
		return false
	}

	// Primeiro verificar compatibilidade genérica
	if areGenericTypesCompatible(target, source) {
		return true
	}

	// Ambos são wrappers de parser.Type
	if wTarget, ok := target.(*ParserTypeWrapper); ok {
		if wSource, ok := source.(*ParserTypeWrapper); ok {
			// Verificar se algum é UnionType
			if unionTarget, ok := wTarget.Type.(*parser.UnionType); ok {
				// Target é união, verificar se source é compatível com algum dos tipos
				for _, typ := range unionTarget.Types {
					if AreParserTypesCompatible(typ, wSource.Type) {
						return true
					}
				}
				return false
			}

			if unionSource, ok := wSource.Type.(*parser.UnionType); ok {
				// Source é união, verificar se target é compatível com algum dos tipos
				for _, typ := range unionSource.Types {
					if AreParserTypesCompatible(wTarget.Type, typ) {
						return true
					}
				}
				return false
			}

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
				if !AreTypesCompatible(mTarget.Types[i], mSource.Types[i]) {
					return false
				}
			}
			return true
		}
		// Tentar comparar MultiValueType de um único elemento com tipo simples
		if len(mTarget.Types) == 1 {
			return AreTypesCompatible(mTarget.Types[0], source)
		}
		return false
	}

	// Tentar comparar tipo simples com MultiValueType de um único elemento
	if mSource, ok := source.(*MultiValueType); ok {
		if len(mSource.Types) == 1 {
			return AreTypesCompatible(target, mSource.Types[0])
		}
		return false
	}

	return false
}

// AreParserTypesCompatible verifica compatibilidade entre parser.Types
func AreParserTypesCompatible(target, source parser.Type) bool {
	if target == nil || source == nil {
		return false
	}

	targetStr := StringifyParserType(target)
	sourceStr := StringifyParserType(source)

	if sourceStr == "any" || targetStr == "any" {
		return true
	}

	if targetStr == "error" || sourceStr == "error" {
		return true
	}

	// Se são exatamente iguais
	if targetStr == sourceStr {
		return true
	}

	// Caso especial: compatibilidade de arrays
	if tArr, ok := target.(*parser.ArrayType); ok {
		if sArr, ok := source.(*parser.ArrayType); ok {
			// Permitir any[] para string[] e vice-versa
			if targetStr == "any[]" || sourceStr == "any[]" {
				return true
			}
			// Verificar compatibilidade dos tipos de elemento
			return AreParserTypesCompatible(tArr.ElementType, sArr.ElementType)
		}
		// Permitir any[] para qualquer array
		if targetStr == "any[]" && strings.HasSuffix(sourceStr, "[]") {
			return true
		}
		// Permitir qualquer array para any[]
		if sourceStr == "any[]" && strings.HasSuffix(targetStr, "[]") {
			return true
		}
	}

	// Caso especial para spread arrays: any[] pode ser atribuído a string[]
	if targetStr == "string[]" && sourceStr == "any[]" {
		return true
	}

	// 1. Target é Identificador (Raw Type) e Source é Genérico (Especializado)
	// Ex: var c Car = generic<string> Car {...}
	if tIdent, ok := target.(*parser.IdentifierType); ok {
		if sGeneric, ok := source.(*parser.GenericType); ok {
			return tIdent.Name == sGeneric.Name
		}
	}

	// 2. Target é Genérico e Source é Identificador (Raw Type)
	// Menos comum, mas possível se inicializar sem generics explícitos
	if tGeneric, ok := target.(*parser.GenericType); ok {
		if sIdent, ok := source.(*parser.IdentifierType); ok {
			return tGeneric.Name == sIdent.Name
		}
	}

	// 3. Comparação Genérico vs Genérico
	if tGeneric, ok := target.(*parser.GenericType); ok {
		if sGeneric, ok := source.(*parser.GenericType); ok {
			if tGeneric.Name != sGeneric.Name {
				return false
			}
			// Se target não tem argumentos (raw), aceita
			if len(tGeneric.TypeArgs) == 0 {
				return true
			}
			if len(tGeneric.TypeArgs) != len(sGeneric.TypeArgs) {
				return false
			}
			for i := range tGeneric.TypeArgs {
				if !AreParserTypesCompatible(tGeneric.TypeArgs[i], sGeneric.TypeArgs[i]) {
					return false
				}
			}
			return true
		}
	}
	// =====================================================

	// Se target for UnionType
	if unionTarget, ok := target.(*parser.UnionType); ok {
		for _, typ := range unionTarget.Types {
			if AreParserTypesCompatible(typ, source) {
				return true
			}
		}
		return false
	}

	// Se source for UnionType
	if unionSource, ok := source.(*parser.UnionType); ok {
		for _, typ := range unionSource.Types {
			if AreParserTypesCompatible(target, typ) {
				return true
			}
		}
		return false
	}

	// Conversões numéricas implícitas e nullables
	if targetStr == "float" && sourceStr == "int" {
		return true
	}
	if targetStr == "int?" && sourceStr == "int" {
		return true
	}
	if targetStr == "int" && sourceStr == "int?" {
		return true
	}
	if targetStr == "float?" && sourceStr == "float" {
		return true
	}
	if targetStr == "float" && sourceStr == "float?" {
		return true
	}
	if targetStr == "bool?" && sourceStr == "bool" {
		return true
	}
	if targetStr == "bool" && sourceStr == "bool?" {
		return true
	}

	return false
}

// =================================================================
// FUNÇÕES ORIGINAIS (AGORA USANDO NOSSA INTERFACE)
// =================================================================

// StringifyType converte nosso Type em string
func StringifyType(t Type) string {
	if t == nil {
		return "void"
	}
	return t.String()
}

// isGenericTypeName verifica se um nome de tipo é genérico (letra maiúscula)
func IsGenericTypeName(name string) bool {
	if len(name) == 1 {
		ch := name[0]
		return ch >= 'A' && ch <= 'Z'
	}
	return false
}

// GetBaseTypeName obtém o nome base de um tipo (remove modificadores)
func GetBaseTypeName(t Type) string {
	if t == nil {
		return ""
	}

	typeStr := StringifyType(t)

	// Remover modificadores: ?, *, []
	if strings.HasSuffix(typeStr, "?") {
		typeStr = typeStr[:len(typeStr)-1]
	}
	if strings.HasPrefix(typeStr, "*") {
		typeStr = typeStr[1:]
	}
	if strings.HasSuffix(typeStr, "[]") {
		typeStr = typeStr[:len(typeStr)-2]
	}

	return typeStr
}

// areGenericTypesCompatible verifica compatibilidade quando envolvem tipos genéricos
func areGenericTypesCompatible(target, source Type) bool {
	targetStr := StringifyType(target)
	sourceStr := StringifyType(source)

	// Se algum for genérico, permitir mais flexibilidade
	if IsGenericTypeName(targetStr) || IsGenericTypeName(sourceStr) {
		// Para operações aritméticas, genéricos podem ser compatíveis com numéricos
		if targetStr == "int" || targetStr == "float" {
			return sourceStr == "int" || sourceStr == "float" || IsGenericTypeName(sourceStr)
		}
		if sourceStr == "int" || sourceStr == "string" {
			return targetStr == "int" || targetStr == "float" || IsGenericTypeName(targetStr)
		}
		// Dois genéricos são sempre compatíveis (serão resolvidos na instanciação)
		if IsGenericTypeName(targetStr) && IsGenericTypeName(sourceStr) {
			return true
		}
	}

	return false
}
