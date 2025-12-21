package lexer

func (s *Scanner) NextToken() Token {
	s.skipSpaceAndComments()

	if s.isEOF() {
		return Token{Type: EOF, Line: s.line, Col: s.col}
	}

	s.start = s.index
	s.tokenLine = s.line
	s.tokenCol = s.col

	ch := s.peek(0)

	switch {
	case isLetter(ch) || ch == '_':
		return s.lexIdentifier()
	case isDigit(ch):
		return s.lexNumber()
	case ch == '"':
		return s.lexString()
	default:
		return s.lexOperator()
	}
}

// skipSpaceAndComments ignora espaços e comentários (// e /* */)
func (s *Scanner) skipSpaceAndComments() {
	for !s.isEOF() {
		ch := s.peek(0)

		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			s.advance()
		case ch == '/' && s.peek(1) == '/':
			s.skipLineComment()
		case ch == '/' && s.peek(1) == '*':
			s.skipBlockComment()
		default:
			return
		}
	}
}

func (s *Scanner) skipLineComment() {
	// Avança "//"
	s.advance()
	s.advance()

	// Avança até o fim da linha
	for !s.isEOF() && s.peek(0) != '\n' {
		s.advance()
	}
}

func (s *Scanner) skipBlockComment() {
	// Avança "/*"
	s.advance()
	s.advance()

	for !s.isEOF() {
		if s.peek(0) == '*' && s.peek(1) == '/' {
			// Avança "*/"
			s.advance()
			s.advance()
			return
		}
		s.advance()
	}
}
