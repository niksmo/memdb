package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

const (
	setOpArgsLen = 2
	getOpArgsLen = 1
	delOpArgsLen = 1

	minArgs = 2
	maxArgs = 3

	keyIdx     = 0
	payloadIdx = 1
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(payload []byte) (domain.Operation, error) {
	if len(payload) == 0 {
		return domain.Operation{}, errors.New("payload is empty")
	}

	args := strings.Fields(string(payload))

	n := len(args)
	if n < minArgs || n > maxArgs {
		return domain.Operation{}, errors.New("invalid number of arguments")
	}

	opCode, err := domain.ParseOpCode(args[0])
	if err != nil {
		return domain.Operation{}, err
	}

	operationArgs := args[1:]

	if err = p.validate(opCode, operationArgs); err != nil {
		return domain.Operation{}, err
	}

	return p.buildOperation(opCode, operationArgs)
}

func (p *Parser) validate(opCode domain.OpCode, operationArgs []string) error {
	const msgFormat = "invalid arguments count, expected %d got %d"

	argsLen := len(operationArgs)
	switch opCode {
	case domain.OpSet:
		if argsLen != setOpArgsLen {
			return fmt.Errorf(msgFormat, setOpArgsLen, argsLen)
		}
	case domain.OpGet:
		if argsLen != getOpArgsLen {
			return fmt.Errorf(msgFormat, getOpArgsLen, argsLen)
		}
	case domain.OpDel:
		if argsLen != delOpArgsLen {
			return fmt.Errorf(msgFormat, delOpArgsLen, argsLen)
		}
	default:
		return errors.New("unexpected operation code")
	}

	return nil
}

func (p *Parser) buildOperation(opCode domain.OpCode, cmdArgs []string) (domain.Operation, error) {
	key := cmdArgs[keyIdx]

	var payload string
	if opCode == domain.OpSet {
		payload = cmdArgs[payloadIdx]
	}

	operation := domain.Operation{
		Code:    opCode,
		Key:     key,
		Payload: payload,
	}

	return operation, nil
}
