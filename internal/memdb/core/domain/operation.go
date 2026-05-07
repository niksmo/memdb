package domain

import (
	"fmt"
	"strings"
)

type OpCode uint8

const (
	OpUnknown OpCode = iota
	OpSet
	OpGet
	OpDel
)

func ParseOpCode(v string) (OpCode, error) {
	switch strings.ToUpper(v) {
	case "SET":
		return OpSet, nil
	case "GET":
		return OpGet, nil
	case "DEL":
		return OpDel, nil
	default:
		return OpUnknown, fmt.Errorf("unsupported operation: %q", v)
	}
}

type Operation struct {
	Code    OpCode
	Key     string
	Payload string
}
