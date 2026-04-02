package parser

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(stmt []string) (cmd, key, value string) {
	cmd, key = stmt[0], stmt[1]

	if len(stmt) == 2 {
		value = stmt[2]
	}

	return
}
