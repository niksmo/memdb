package compute

import (
	"context"

	"github.com/niksmo/memdb/internal/core/models"
)

type Parser interface {
	Parse(ctx context.Context, stmt []byte) (models.Request, error)
}

type Compute struct {
	p Parser
}

func New(p Parser) *Compute {
	return &Compute{p}
}

func (c *Compute) Do(ctx context.Context, stmt []byte) (models.Request, error) {
	if err := ctx.Err(); err != nil {
		return models.Request{}, err
	}

	return c.p.Parse(ctx, stmt)
}
