package wal

import "github.com/niksmo/memdb/internal/memdb/core/domain"

type event struct {
	operation domain.Operation
	done      chan<- struct{}
}

func newEvent(opCode domain.OpCode, key, payload string) (e event, done <-chan struct{}) {
	ch := make(chan struct{})

	op := domain.Operation{
		Code:    opCode,
		Key:     key,
		Payload: payload,
	}

	e = event{
		operation: op,
		done:      ch,
	}

	return e, ch
}
