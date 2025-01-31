package main

type parser struct {
	tokens    []string
	collected string
}

func newParser() *parser {
	return &parser{
		tokens:    []string{},
		collected: "",
	}
}

func (p *parser) addToCollected(addition string) {
	p.collected = p.collected + addition
}

func (p *parser) appendToken(token string) {
	p.tokens = append(p.tokens, token)
}

func (p *parser) resetCollected() {
	p.collected = ""
}
