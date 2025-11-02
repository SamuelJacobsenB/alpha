package lexer

// emit constrói token usando start..index (uma única conversão)
func (s *Scanner) emit(t TokenType) Token {
	lex := string(s.src[s.start:s.index])
	tok := Token{Type: t, Lexeme: lex, Line: s.tokenLine, Col: s.tokenCol}

	switch t {
	case STRING:
		// remove aspas; atenção: assume que start..index inclui as aspas
		if s.index-s.start >= 2 {
			tok.Value = unescapeStringBytes(s.src[s.start+1 : s.index-1])
		} else {
			tok.Value = ""
		}
	case INT, FLOAT:
		tok.Value = lex
	default:
		tok.Value = lex
	}
	return tok
}

func (s *Scanner) errorToken(msg string) Token {
	return Token{Type: ERROR, Lexeme: msg, Value: msg, Line: s.line, Col: s.col}
}

// LexAll retorna todos tokens (útil em testes)
func LexAll(src string) []Token {
	sc := NewScanner(src)
	out := make([]Token, 0, 64)
	for {
		t := sc.NextToken()
		out = append(out, t)
		if t.Type == EOF || t.Type == ERROR {
			break
		}
	}
	return out
}
