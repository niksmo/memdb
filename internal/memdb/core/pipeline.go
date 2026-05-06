package core

import (
	"context"

	"github.com/niksmo/memdb/internal/memdb/core/models"
)

type Compute interface {
	Do(ctx context.Context, buf []byte) (models.Request, error)
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

func (p *Pipeline) Handle(ctx context.Context, buf []byte) ([]byte, error) {
	req, err := p.c.Do(ctx, buf)
	if err != nil {
		return nil, err
	}

	resp, err := p.s.Process(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
