package core

import (
	"context"
	"errors"

	"github.com/niksmo/memdb/internal/core/models"
	"github.com/niksmo/memdb/internal/core/storage"
)

var ErrInternal = errors.New("internal error")

type Compute interface {
	Do(ctx context.Context, stmt []byte) (models.Request, error)
}

type Storage interface {
	Process(ctx context.Context, req models.Request) ([]byte, error)
}

type Pipeline struct {
	c Compute
	s Storage
}

func NewPipeline(c Compute, s Storage) *Pipeline {
	return &Pipeline{c, s}
}

func (p *Pipeline) Exec(ctx context.Context, stmt []byte) ([]byte, error) {
	req, err := p.c.Do(ctx, stmt)
	if err != nil {
		return nil, err
	}

	resp, err := p.s.Process(ctx, req)
	if err != nil {
		if errors.Is(err, storage.ErrUnknownCommand) {
			return nil, ErrInternal
		}
		return nil, err
	}

	return resp, nil
}
