package parser

import "strings"

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(stmt []string) (cmd, key, value string) {
	cmd, key = strings.ToUpper(stmt[0]), stmt[1]

	if len(stmt) == 3 {
		value = stmt[2]
	}

	return
}
