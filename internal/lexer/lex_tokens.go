package lexer

func (s *Scanner) lexIdentifier() Token {
	for !s.isEOF() {
		ch := s.peek(0)
		if isLetter(ch) || isDigit(ch) || ch == '_' || ch == '?' {
			s.advance()
			continue
		}
		break
	}

	tok := s.emit(IDENT)
	if _, ok := keywords[tok.Lexeme]; ok {
		tok.Type = KEYWORD
	}
	return tok
}

// lexNumber reconhece INT e FLOAT (suporte básico a expoente)
func (s *Scanner) lexNumber() Token {
	// parte inteira
	for !s.isEOF() && isDigit(s.peek(0)) {
		s.advance()
	}

	// fracional?
	if s.peek(0) == '.' && isDigit(s.peek(1)) {
		s.advance() // advance '.'
		for !s.isEOF() && isDigit(s.peek(0)) {
			s.advance()
		}
		// expoente opcional
		if ch := s.peek(0); ch == 'e' || ch == 'E' {
			s.advance() // advance 'e' or 'E'
			if s.peek(0) == '+' || s.peek(0) == '-' {
				s.advance()
			}
			if !isDigit(s.peek(0)) {
				return s.errorToken("malformed exponent")
			}
			for !s.isEOF() && isDigit(s.peek(0)) {
				s.advance()
			}
		}
		return s.emit(FLOAT)
	}

	// expoente sem ponto (ex: 1e10)
	if ch := s.peek(0); ch == 'e' || ch == 'E' {
		s.advance()
		if s.peek(0) == '+' || s.peek(0) == '-' {
			s.advance()
		}
		if !isDigit(s.peek(0)) {
			return s.errorToken("malformed exponent")
		}
		for !s.isEOF() && isDigit(s.peek(0)) {
			s.advance()
		}
		return s.emit(FLOAT)
	}

	return s.emit(INT)
}

// lexString consome "..." com escapes simples
func (s *Scanner) lexString() Token {
	s.advance() // consome abertura "

	escaped := false
	for !s.isEOF() {
		ch := s.peek(0)
		if escaped {
			escaped = false
			s.advance()
			continue
		}
		if ch == '\\' {
			escaped = true
			s.advance()
			continue
		}
		if ch == '"' {
			s.advance() // consome fechamento
			return s.emit(STRING)
		}
		s.advance()
	}
	return s.errorToken("unterminated string")
}

// lexOperator reconhece operadores 2-char ou fallback 1-char
func (s *Scanner) lexOperator() Token {
	ch := s.peek(0)
	ch1 := s.peek(1)

	// Casos para colchetes (já existe, mas vamos manter)
	if ch == '[' || ch == ']' {
		s.advance()
		return s.emit(OP)
	}

	// Casos para parênteses e chaves
	if ch == '(' || ch == ')' || ch == '{' || ch == '}' {
		s.advance()
		return s.emit(OP)
	}

	// Operadores de 2 caracteres
	twoCharOps := map[string]bool{
		"==": true, "!=": true, "<=": true, ">=": true,
		"&&": true, "||": true, "++": true, "--": true,
		"+=": true, "-=": true, "*=": true, "/=": true,
	}

	if twoCharOps[string(ch)+string(ch1)] {
		s.advance()
		s.advance()
		return s.emit(OP)
	}

	// Operadores de 1 caractere
	oneCharOps := map[byte]bool{
		'+': true, '-': true, '*': true, '/': true,
		'%': true, '<': true, '>': true, '=': true,
		'!': true, '&': true, '|': true, ';': true,
		',': true, '.': true, ':': true,
	}

	if oneCharOps[ch] {
		s.advance()
		return s.emit(OP)
	}

	// Operador não reconhecido
	s.advance()
	return s.errorToken("operador não reconhecido: " + string(ch))
}
