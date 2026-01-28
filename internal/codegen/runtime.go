package codegen

// GetRuntime retorna o código Go padrão que todo arquivo precisa ter
func GetRuntime() string {
	return `
// ================= RUNTIME =================
type GenericMap map[string]interface{}
type GenericSet map[interface{}]bool

// Helper para converter bool em int (usado em condições if c > 0)
func B2I(b bool) int64 {
	if b { return 1 }
	return 0
}

// Builtin: append
func _append(slice interface{}, val interface{}) interface{} {
    // Implementação simplificada usando reflection ou generics do Go 1.18+
    // Aqui assumimos slices de interface{} para simplicidade do exemplo
    return append(slice.([]interface{}), val)
}

// Builtin: length
func _len(v interface{}) int64 {
    switch val := v.(type) {
    case string: return int64(len(val))
    case []interface{}: return int64(len(val))
    case map[string]interface{}: return int64(len(val))
    default: return 0
    }
}
// ===========================================
`
}
