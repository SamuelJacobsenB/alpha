package lexer

import "bytes"

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func unescapeStringBytes(b []byte) string {
	// Verifica se não há escapes para retornar rápido
	if bytes.IndexByte(b, '\\') == -1 {
		return string(b)
	}

	result := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\\' && i+1 < len(b) {
			i++
			result = append(result, escapeMap[b[i]])
		} else {
			result = append(result, b[i])
		}
	}

	return string(result)
}

var escapeMap = [256]byte{
	'n':  '\n',
	't':  '\t',
	'r':  '\r',
	'"':  '"',
	'\\': '\\',
}
