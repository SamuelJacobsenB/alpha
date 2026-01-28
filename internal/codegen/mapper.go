package codegen

import (
	"strings"

	"github.com/alpha/internal/semantic"
)

// ToGoType converte estritamente tipos do Alpha para Go
func ToGoType(t semantic.Type) string {
	if t == nil {
		return "interface{}"
	}

	// Usamos a função que você já definiu no semantic.txt para pegar a string do tipo
	typeStr := semantic.StringifyType(t)

	if strings.HasSuffix(typeStr, "?") {
		base := strings.TrimSuffix(typeStr, "?")
		// Retorna ponteiro: int? -> *int64
		return "*" + mapSimpleString(base)
	}

	// Limpeza para tipos que podem vir com ponteiros ou opcionais do seu parser
	typeStr = strings.TrimPrefix(typeStr, "*")
	typeStr = strings.TrimSuffix(typeStr, "?")

	switch typeStr {
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
	default:
		// Caso seja um array: int[] -> []int64
		if strings.HasSuffix(typeStr, "[]") {
			elem := strings.TrimSuffix(typeStr, "[]")
			return "[]" + mapSimpleString(elem)
		}

		// Se não for primitivo, é o nome da Struct/Custom type
		return typeStr
	}
}

func mapSimpleString(s string) string {
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
	default:
		// Se não for primitivo, assume que é uma Struct ou Interface definida pelo usuário
		return s
	}
}

// ZeroValue retorna o valor inicial para declaração de variáveis
func ZeroValue(goType string) string {
	switch goType {
	case "int64":
		return "0"
	case "float64":
		return "0.0"
	case "bool":
		return "false"
	case "string":
		return "\"\""
	case "interface{}":
		return "nil"
	default:
		if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "*") {
			return "nil"
		}
		return goType + "{}" // Struct vazia
	}
}
