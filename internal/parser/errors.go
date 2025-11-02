package parser

func (p *Parser) HasErrors() bool { return len(p.Errors) > 0 }
func (p *Parser) ErrorsText() string {
	// simples join
	if len(p.Errors) == 0 {
		return ""
	}
	out := ""
	for _, e := range p.Errors {
		out += e + "\n"
	}
	return out
}
