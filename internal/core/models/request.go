package models

import (
	"errors"
)

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrEmptyKey       = errors.New("empty key")
	ErrEmptyValue     = errors.New("value is empty")
	ErrMustEmptyValue = errors.New("value must be empty")
)

type Request struct {
	cmd   Command
	key   string
	value string
}

func NewRequest(cmd Command, key string, value string) (Request, error) {
	if key == "" {
		return Request{}, ErrEmptyKey
	}

	switch cmd {
	case CommandSet:
		if value == "" {
			return Request{}, ErrEmptyValue
		}
	case CommandGet, CommandDel:
		if value != "" {
			return Request{}, ErrMustEmptyValue
		}
	default:
		return Request{}, ErrInvalidCommand
	}

	r := Request{
		cmd:   cmd,
		key:   key,
		value: value,
	}

	return r, nil
}

func (r Request) Cmd() Command {
	return r.cmd
}

func (r Request) Key() string {
	return r.key
}

func (r Request) Value() string {
	return r.value
}
