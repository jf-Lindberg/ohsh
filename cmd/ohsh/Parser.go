package ohsh

type SymbolMode int

const (
	Normal SymbolMode = iota
	SingleQuote
	DoubleQuote
	Escaped
)

type Parser struct {
	tokens             []string
	currentToken       string
	position           int
	tokenStart         int
	symbolMode         SymbolMode
	previousSymbolMode SymbolMode
}

func NewParser() *Parser {
	return &Parser{
		tokens:             []string{},
		currentToken:       "",
		position:           0,
		tokenStart:         0,
		symbolMode:         Normal,
		previousSymbolMode: Normal,
	}
}

func (p *Parser) ParseLine(line string) (Command, error) {
	for p.position < len(line) {
		r := line[p.position]
		switch p.symbolMode {
		case Normal:
			if err := p.handleNormalState(r, line, redirector); err != nil {
				return err
			}
		case SingleQuote:
			p.handleSingleQuoteState(r, input)
		case DoubleQuote:
			p.handleDoubleQuoteState(r, input)
		case Escaped:
			p.handleEscapedState(r, input)
		}
	}

	if p.tokenStart < len(input) {
		p.appendToken(input[p.tokenStart:])
	}

	return nil
}

func (p *Parser) Reset() {
	// This probably doesn't work
	p = NewParser()
}
