package parser

type Parser interface {
	// Parse reads a transmission from a player and returns a structured representation of the transmitted request.
	// The type of the request must be determined by reflection.
	Parse(string) any
}

type parser struct{}

func New() Parser {
	return &parser{}
}

func (p *parser) Parse(s string) any {
	return nil
}
