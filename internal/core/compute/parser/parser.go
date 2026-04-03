package parser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/niksmo/memdb/internal/core/models"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(_ context.Context, stmt []byte) (models.Request, error) {
	if err := p.validateStmt(stmt); err != nil {
		return models.Request{}, err
	}

	args := bytes.Fields(stmt)

	if err := p.validateArgs(args); err != nil {
		return models.Request{}, err
	}

	return p.parseRequest(args)
}

func (p *Parser) validateStmt(stmt []byte) error {
	if len(stmt) == 0 {
		return errors.New("statement is empty")
	}

	if !utf8.Valid(stmt) {
		return errors.New("statement should be the valid encoded string")
	}

	return nil
}

func (p *Parser) validateArgs(args [][]byte) error {
	n := len(args)

	if !(n == 2 || n == 3) {
		return errors.New("invalid number of statement arguments")
	}

	return nil
}

func (p *Parser) isValue(args [][]byte) bool {
	return len(args) == 3
}

func (p *Parser) parseRequest(args [][]byte) (models.Request, error) {
	var (
		cmd        models.Command
		key, value string
	)

	cmd, err := models.ParseCommand(string(args[0]))
	if err != nil {
		return models.Request{}, errors.New("invalid statement command")
	}

	key = string(args[1])

	if p.isValue(args) {
		value = string(args[2])
	}

	req, err := models.NewRequest(cmd, key, value)
	if err != nil {
		return models.Request{}, fmt.Errorf("invalid statement: %w", err)
	}

	return req, err
}
