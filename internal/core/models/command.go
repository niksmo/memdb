package models

import "fmt"

type Command uint8

const (
	CommandUnknown Command = iota
	CommandSet
	CommandGet
	CommandDel
)

func ParseCommand(v string) (Command, error) {
	switch v {
	case "SET":
		return CommandSet, nil
	case "GET":
		return CommandGet, nil
	case "DEL":
		return CommandDel, nil
	default:
		return CommandUnknown, fmt.Errorf("invalid command %q", v)
	}
}
