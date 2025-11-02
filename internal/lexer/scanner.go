package lexer

// Scanner simples e performático usando []byte
type Scanner struct {
	src       []byte
	n         int // length of src
	index     int // próximo byte a ler
	start     int // início do token corrente
	line      int // linha atual (1-based)
	col       int // coluna atual (1-based)
	tokenLine int // linha onde token corrente começou (mantido para emitir)
	tokenCol  int // coluna onde token corrente começou
}

// NewScanner cria scanner a partir de string fonte
func NewScanner(src string) *Scanner {
	b := []byte(src)
	return &Scanner{src: b, n: len(b), index: 0, start: 0, line: 1, col: 1}
}

func (s *Scanner) isEOF() bool { return s.index >= s.n }

func (s *Scanner) peek(off int) byte {
	i := s.index + off
	if i >= s.n || i < 0 {
		return 0
	}
	return s.src[i]
}

func (s *Scanner) advance() byte {
	if s.isEOF() {
		return 0
	}
	ch := s.src[s.index]
	s.index++
	if ch == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	return ch
}
