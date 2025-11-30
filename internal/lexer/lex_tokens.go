package lexer

func (s *Scanner) lexIdentifier() Token {
	for !s.isEOF() {
		ch := s.peek(0)
		if isLetter(ch) || isDigit(ch) || ch == '_' {
			s.advance()
			continue
		}
		// Permitir '?' apenas no final (para tipos opcionais)
		if ch == '?' && (s.index-s.start > 0) {
			s.advance()
			break
		}
		break
	}

	lex := string(s.src[s.start:s.index])

	if _, ok := keywords[lex]; ok {
		return s.emit(KEYWORD)
	}

	// Detecção de genéricos (T, U, etc.)
	if len(lex) == 1 && lex[0] >= 'A' && lex[0] <= 'Z' {
		return s.emit(GENERIC)
	}

	return s.emit(IDENT)
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

	// CORREÇÃO: Garantir que o scanner avançou completamente sobre o número
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

func (s *Scanner) lexOperator() Token {
	ch := s.peek(0)
	ch1 := s.peek(1)

	// Operadores de 2 caracteres - VERIFICAR PRIMEIRO
	twoCharOps := map[string]bool{
		"==": true, "!=": true, "<=": true, ">=": true,
		"&&": true, "||": true, "++": true, "--": true,
		"+=": true, "-=": true, "*=": true, "/=": true,
	}

	twoChar := string(ch) + string(ch1)
	if twoCharOps[twoChar] {
		s.advance()
		s.advance()
		return s.emit(OP)
	}

	// Operadores de 1 caractere - SIMPLIFICADO
	oneCharOps := map[byte]bool{
		'+': true, '-': true, '*': true, '/': true,
		'%': true, '=': true, '!': true, '&': true,
		'|': true, ';': true, ',': true, '.': true,
		':': true, '(': true, ')': true, '{': true,
		'}': true, '[': true, ']': true, '<': true,
		'>': true,
	}

	if oneCharOps[ch] {
		s.advance()
		return s.emit(OP)
	}

	// Se chegou aqui, é um operador não reconhecido
	s.advance()
	return s.errorToken("operador não reconhecido: " + string(ch))
}
