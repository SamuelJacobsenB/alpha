package lexer

// Scanner []byte
type Scanner struct {
	src           []byte
	index         int       // próximo byte a ler
	start         int       // início do token corrente
	line          int       // linha atual (1-based)
	col           int       // coluna atual (1-based)
	tokenLine     int       // linha onde token corrente começou
	tokenCol      int       // coluna onde token corrente começou
	lastTokenType TokenType // último token emitido
}

func NewScanner(src string) *Scanner {
	b := []byte(src)
	return &Scanner{
		src:   b,
		index: 0,
		line:  1,
		col:   1,
	}
}

func (s *Scanner) isEOF() bool { return s.index >= len(s.src) }

func (s *Scanner) peek(off int) byte {
	i := s.index + off
	if i >= 0 && i < len(s.src) {
		return s.src[i]
	}
	return 0
}

func (s *Scanner) advance() {
	if s.isEOF() {
		return
	}

	ch := s.src[s.index]
	s.index++

	switch ch {
	case '\n':
		s.line++
		s.col = 1
	case '\r':
		// Tratar \r\n como uma única quebra de linha
		if s.index < len(s.src) && s.src[s.index] == '\n' {
			s.index++
		}
		s.line++
		s.col = 1
	default:
		s.col++
	}
}
