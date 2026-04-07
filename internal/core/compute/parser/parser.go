package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/niksmo/memdb/internal/core/models"
)

const (
	setCmdArgsLen = 2
	getCmdArgsLen = 1
	delCmdArgsLen = 1

	minArgs = 2
	maxArgs = 3

	keyIdx   = 0
	valueIdx = 1
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(stmt []byte) (models.Request, error) {
	if len(stmt) == 0 {
		return models.Request{}, errors.New("statement is empty")
	}

	args := strings.Fields(string(stmt))

	n := len(args)
	if n < minArgs || n > maxArgs {
		return models.Request{}, errors.New("invalid number of statement arguments")
	}

	cmd, err := models.ParseCommand(args[0])
	if err != nil {
		return models.Request{}, fmt.Errorf("parse command: %w", err)
	}

	cmdArgs := args[1:]

	if err = p.validate(cmd, cmdArgs); err != nil {
		return models.Request{}, err
	}

	return p.buildRequest(cmd, cmdArgs)
}

func (p *Parser) validate(cmd models.Command, cmdArgs []string) error {
	const msgFormat = "invalid argument count, expected %d got %d"

	n := len(cmdArgs)
	switch cmd {
	case models.CommandSet:
		if n != setCmdArgsLen {
			return fmt.Errorf(msgFormat, setCmdArgsLen, n)
		}
	case models.CommandGet:
		if n != getCmdArgsLen {
			return fmt.Errorf(msgFormat, getCmdArgsLen, n)
		}
	case models.CommandDel:
		if n != delCmdArgsLen {
			return fmt.Errorf(msgFormat, delCmdArgsLen, n)
		}
	default:
		return errors.New("unexpected command")
	}

	return nil
}

func (p *Parser) buildRequest(cmd models.Command, cmdArgs []string) (models.Request, error) {
	key := cmdArgs[keyIdx]

	var value string
	if cmd == models.CommandSet {
		value = cmdArgs[valueIdx]
	}

	req := models.Request{
		Cmd:   cmd,
		Key:   key,
		Value: value,
	}

	return req, nil
}
