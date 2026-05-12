package compute

import (
	"context"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

type Parser interface {
	Parse(payload []byte) (domain.Operation, error)
}

type Compute struct {
	p Parser
}

func New(p Parser) *Compute {
	return &Compute{p}
}

func (c *Compute) Do(ctx context.Context, payload []byte) (domain.Operation, error) {
	if err := ctx.Err(); err != nil {
		return domain.Operation{}, err
	}

	return c.p.Parse(payload)
}
