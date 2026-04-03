package core

import (
	"context"

	"github.com/niksmo/memdb/internal/core/models"
)

type Compute interface {
	Do(ctx context.Context, stmt []byte) (models.Request, error)
}

type Pipeline struct {
	c Compute
}

func NewPipeline(c Compute) *Pipeline {
	return &Pipeline{c}
}

func (p *Pipeline) Exec(ctx context.Context, stmt []byte) ([]byte, error) {
	q, err := p.c.Do(ctx, stmt)
	if err != nil {
		return nil, err
	}
	_ = q

	return nil, err
}
