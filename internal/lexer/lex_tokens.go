package lexer

var (
	twoCharOps = map[string]bool{
		"==": true, "!=": true, "<=": true, ">=": true,
		"&&": true, "||": true, "++": true, "--": true,
		"+=": true, "-=": true, "*=": true, "/=": true,
	}

	oneCharOps = map[byte]bool{
		'+': true, '-': true, '*': true, '/': true,
		'%': true, '=': true, '!': true, '&': true,
		'|': true, ';': true, ',': true, '.': true,
		':': true, '(': true, ')': true, '{': true,
		'}': true, '[': true, ']': true, '<': true,
		'>': true,
	}
)

func (s *Scanner) lexIdentifier() Token {
	// Consome letras, dígitos e underscore
	for !s.isEOF() && (isLetter(s.peek(0)) || isDigit(s.peek(0)) || s.peek(0) == '_') {
		s.advance()
	}

	// Permite '?' apenas no final
	if !s.isEOF() && s.peek(0) == '?' && s.index > s.start {
		s.advance()
	}

	lex := string(s.src[s.start:s.index])

	// Verifica se é keyword
	if _, isKeyword := keywords[lex]; isKeyword {
		return s.emit(KEYWORD)
	}

	// Verifica se é tipo genérico (letra maiúscula única)
	if len(lex) == 1 && lex[0] >= 'A' && lex[0] <= 'Z' {
		return s.emit(GENERIC)
	}

	return s.emit(IDENT)
}

func (s *Scanner) lexNumber() Token {
	// Parte inteira
	for !s.isEOF() && isDigit(s.peek(0)) {
		s.advance()
	}

	// Verifica se tem parte decimal
	hasDecimal := false
	if s.peek(0) == '.' && isDigit(s.peek(1)) {
		s.advance() // consume '.'
		hasDecimal = true
		for !s.isEOF() && isDigit(s.peek(0)) {
			s.advance()
		}
	}

	// Verifica expoente
	if ch := s.peek(0); ch == 'e' || ch == 'E' {
		if !s.parseExponent() {
			return s.errorToken("malformed exponent")
		}
		hasDecimal = true
	}

	if hasDecimal {
		return s.emit(FLOAT)
	}
	return s.emit(INT)
}

func (s *Scanner) parseExponent() bool {
	s.advance() // consume 'e' or 'E'

	// Sinal opcional
	if s.peek(0) == '+' || s.peek(0) == '-' {
		s.advance()
	}

	// Deve ter pelo menos um dígito
	if !isDigit(s.peek(0)) {
		return false
	}

	// Consome dígitos do expoente
	for !s.isEOF() && isDigit(s.peek(0)) {
		s.advance()
	}

	return true
}

func (s *Scanner) lexString() Token {
	s.advance() // consume opening "

	for !s.isEOF() {
		ch := s.peek(0)
		s.advance()

		if ch == '\\' && !s.isEOF() {
			s.advance() // skip escaped character
		} else if ch == '"' {
			return s.emit(STRING)
		}
	}

	return s.errorToken("unterminated string")
}

func (s *Scanner) lexOperator() Token {
	ch1 := s.peek(0)
	ch2 := s.peek(1)

	// Verifica operadores de 2 caracteres
	if ch2 != 0 && twoCharOps[string(ch1)+string(ch2)] {
		s.advance()
		s.advance()
		return s.emit(OP)
	}

	// Verifica operadores de 1 caractere
	if oneCharOps[ch1] {
		// Verificação especial para determinar se é um tipo genérico
		if ch1 == '<' && s.isLikelyGeneric() {
			// Trata como delimitador de tipo genérico
			s.advance()
			return s.emit(OP) // Continua sendo OP, mas o parser saberá pelo contexto
		}
		s.advance()
		return s.emit(OP)
	}

	// Operador não reconhecido
	s.advance()
	return s.errorToken("operador não reconhecido: " + string(ch1))
}

// isLikelyGeneric verifica se o próximo '<' provavelmente inicia um tipo genérico
func (s *Scanner) isLikelyGeneric() bool {
	// Olha para trás para verificar se há uma palavra-chave que indica genéricos
	if s.index == 0 {
		return false
	}

	// Encontra o início do token anterior
	i := s.index - 1
	// Retrocede enquanto encontra espaços ou quebras de linha
	for i >= 0 && (s.src[i] == ' ' || s.src[i] == '\t' || s.src[i] == '\n' || s.src[i] == '\r') {
		i--
	}

	if i < 0 {
		return false
	}

	// Encontra o início da palavra anterior
	start := i
	for start >= 0 && (isLetter(s.src[start]) || isDigit(s.src[start]) || s.src[start] == '_') {
		start--
	}
	start++ // Ajusta para o primeiro caractere da palavra

	if start > i {
		return false
	}

	// Verifica se a palavra anterior é "generic" ou uma letra maiúscula única (tipo genérico)
	prevWord := string(s.src[start : i+1])

	// Se for "generic", é definitivamente um tipo genérico
	if prevWord == "generic" {
		return true
	}

	// Se for uma letra maiúscula única, também pode ser parte de um tipo genérico
	if len(prevWord) == 1 && prevWord[0] >= 'A' && prevWord[0] <= 'Z' {
		return true
	}

	return false
}
