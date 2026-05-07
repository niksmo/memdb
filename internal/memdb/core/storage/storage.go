package storage

import (
	"context"
	"errors"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

type Engine interface {
	Set(key string, payload string)
	Get(key string) (payload string, err error)
	Del(key string)
}

type WAL interface {
	Run() error
	WriteLog(ctx context.Context, opCode domain.OpCode, key, payload string) error
	LoadAll(ctx context.Context, cb func(opCode domain.OpCode, key, payload string)) error
	Close() error
}

var okResponse = []byte("OK")

var walCommands = map[domain.OpCode]struct{}{
	domain.OpSet: {},
	domain.OpDel: {},
}

type Options struct {
	Engine     Engine
	WAL        WAL
	WALEnabled bool
}

type Storage struct {
	engine     Engine
	wal        WAL
	walEnabled bool
}

func New(o Options) *Storage {
	return &Storage{
		engine:     o.Engine,
		wal:        o.WAL,
		walEnabled: o.WALEnabled,
	}
}

func (s *Storage) Run() error {
	if !s.walEnabled {
		return nil
	}

	return s.wal.Run()
}

func (s *Storage) Process(ctx context.Context, operation domain.Operation) ([]byte, error) {
	var (
		opCode  = operation.Code
		key     = operation.Key
		payload = operation.Payload
	)

	if err := s.writeLog(ctx, opCode, key, payload); err != nil {
		return nil, err
	}

	switch opCode {
	case domain.OpSet:
		s.engine.Set(key, payload)
		return okResponse, nil

	case domain.OpGet:
		data, err := s.engine.Get(key)
		if err != nil {
			return nil, err
		}
		return []byte(data), nil

	case domain.OpDel:
		s.engine.Del(key)
		return okResponse, nil

	default:
		return nil, errors.New("unexpected operation")
	}
}

func (s *Storage) Load(ctx context.Context) error {
	if !s.walEnabled {
		return nil
	}

	return s.wal.LoadAll(ctx, func(opCode domain.OpCode, key, payload string) {
		if opCode == domain.OpSet {
			s.engine.Set(key, payload)
			return
		}

		if opCode == domain.OpDel {
			s.engine.Del(key)
			return
		}
	})
}

func (s *Storage) Close() error {
	if !s.walEnabled {
		return nil
	}

	return s.wal.Close()
}

func (s *Storage) writeLog(ctx context.Context, opCode domain.OpCode, key, payload string) error {
	if !s.walEnabled {
		return nil
	}

	if _, ok := walCommands[opCode]; !ok {
		return nil
	}

	return s.wal.WriteLog(ctx, opCode, key, payload)
}
