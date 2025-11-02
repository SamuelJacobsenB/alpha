package lexer

import "bytes"

// unescapeStringBytes processa escapes básicos a partir de []byte
func unescapeStringBytes(bs []byte) string {
	var out bytes.Buffer
	for i := 0; i < len(bs); i++ {
		b := bs[i]
		if b != '\\' {
			out.WriteByte(b)
			continue
		}
		// escape
		i++
		if i >= len(bs) {
			out.WriteByte('\\')
			break
		}
		switch bs[i] {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case '"':
			out.WriteByte('"')
		case '\\':
			out.WriteByte('\\')
		default:
			out.WriteByte(bs[i])
		}
	}
	return out.String()
}

// helpers simples (ASCII-fast)
func isLetter(b byte) bool { return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') }
func isDigit(b byte) bool  { return b >= '0' && b <= '9' }
