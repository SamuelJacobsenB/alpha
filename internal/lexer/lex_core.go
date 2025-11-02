package lexer

// NextToken: loop principal, pula espaços/comentários e escolhe o lexador
func (s *Scanner) NextToken() Token {
	for {
		s.skipSpaceAndComments()
		if s.isEOF() {
			return Token{Type: EOF, Lexeme: "", Line: s.line, Col: s.col}
		}
		// marcar início do token e guardar posição humana
		s.start = s.index
		s.tokenLine = s.line
		s.tokenCol = s.col

		ch := s.peek(0)

		// identificador ou keyword (letters, underscore, digits depois)
		if isLetter(ch) || ch == '_' {
			return s.lexIdentifier()
		}
		// número
		if isDigit(ch) {
			return s.lexNumber()
		}
		// string
		if ch == '"' {
			return s.lexString()
		}
		// operadores / separadores
		return s.lexOperator()
	}
}

// skipSpaceAndComments ignora espaços e comentários (// e /* */)
func (s *Scanner) skipSpaceAndComments() {
	for !s.isEOF() {
		ch := s.peek(0)
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			s.advance()
			continue
		}

		// // line comment
		if ch == '/' && s.peek(1) == '/' {
			// advance "//"
			s.advance()
			s.advance()

			// advance rest of line
			for !s.isEOF() && s.peek(0) != '\n' {
				s.advance()
			}
			continue
		}

		// /* block comment */
		if ch == '/' && s.peek(1) == '*' {
			// advance "/*"
			s.advance()
			s.advance()

			for !s.isEOF() {
				if s.peek(0) == '*' && s.peek(1) == '/' {
					// advance "*/"
					s.advance()
					s.advance()
					break
				}
				s.advance()
			}
			continue
		}
		break
	}
}
