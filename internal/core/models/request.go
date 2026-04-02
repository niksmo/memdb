package models

import (
	"errors"
)

var (
	ErrInvalidCommand = errors.New("invalid command")
	ErrEmptyKey       = errors.New("empty key")
	ErrInvalidValue   = errors.New("invalid value")
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
			return Request{}, ErrInvalidValue
		}
	case CommandGet, CommandDel:
		if value != "" {
			return Request{}, ErrInvalidValue
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

func (r Request) Valid() bool {
	return r != Request{}
}
