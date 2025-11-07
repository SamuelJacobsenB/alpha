package lexer

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func unescapeStringBytes(b []byte) string {
	// Implementação simples - remove escapes básicos
	// Você pode expandir isso conforme necessário
	var result []byte
	escaped := false

	for i := 0; i < len(b); i++ {
		if escaped {
			switch b[i] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			default:
				result = append(result, b[i])
			}
			escaped = false
		} else if b[i] == '\\' {
			escaped = true
		} else {
			result = append(result, b[i])
		}
	}

	return string(result)
}
