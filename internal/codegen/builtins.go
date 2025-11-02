package codegen

import (
	"fmt"
	"strings"
)

// builtinPrint escreve os valores concatenados com espaço e nova linha
func builtinPrint(args []Value) (Value, error) {
	out := []string{}
	for _, a := range args {
		out = append(out, formatValue(a))
	}
	fmt.Println(strings.Join(out, " "))
	return nil, nil
}

func formatValue(v Value) string {
	if v == nil {
		return "null"
	}
	switch x := v.(type) {
	case int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return fmt.Sprintf("%v", x)
	case string:
		return x
	case bool:
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
