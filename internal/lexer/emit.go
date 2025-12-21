package lexer

func (s *Scanner) emit(t TokenType) Token {
	tok := Token{
		Type:   t,
		Lexeme: string(s.src[s.start:s.index]),
		Line:   s.tokenLine,
		Col:    s.tokenCol,
	}

	if t == STRING && s.index-s.start >= 2 {
		tok.Value = unescapeStringBytes(s.src[s.start+1 : s.index-1])
	} else {
		tok.Value = tok.Lexeme
	}

	s.lastTokenType = t
	return tok
}

func (s *Scanner) errorToken(msg string) Token {
	return Token{
		Type:   ERROR,
		Lexeme: msg,
		Value:  msg,
		Line:   s.line,
		Col:    s.col,
	}
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
