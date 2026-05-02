package storage

import (
	"context"
	"errors"

	"github.com/niksmo/memdb/internal/memdb/core/models"
)

var okResponse = []byte("OK")

type Engine interface {
	Set(key string, value string)
	Get(key string) (value string, err error)
	Del(key string)
}

type Storage struct {
	engine Engine
}

func New(e Engine) *Storage {
	return &Storage{e}
}

func (s *Storage) Process(ctx context.Context, req models.Request) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cmd := req.Cmd
	key := req.Key
	switch cmd {
	case models.CommandSet:
		s.engine.Set(key, req.Value)
		return okResponse, nil

	case models.CommandGet:
		v, err := s.engine.Get(key)
		if err != nil {
			return nil, err
		}
		return []byte(v), nil

	case models.CommandDel:
		s.engine.Del(key)
		return okResponse, nil

	default:
		return nil, errors.New("unexpected command")
	}
}
