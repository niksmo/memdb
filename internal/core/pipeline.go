package core

import (
	"context"

	"github.com/niksmo/memdb/internal/core/models"
)

type Compute interface {
	Do(ctx context.Context, stmt []string) (models.Request, error)
}

type Pipeline struct {
	c Compute
}

func NewPipeline(c Compute) *Pipeline {
	return &Pipeline{c}
}

func (p *Pipeline) Exec(ctx context.Context, stmt []string) (string, error) {
	q, err := p.c.Do(ctx, stmt)
	if err != nil {
		return "", err
	}
	_ = q

	return "", err
}
