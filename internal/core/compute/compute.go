package compute

import (
	"context"

	"github.com/niksmo/memdb/internal/core/models"
)

type Parser interface {
	Parse(stmt []string) (cmd, key, value string)
}

type Compute struct {
	p Parser
}

func New(p Parser) *Compute {
	return &Compute{p}
}

func (c *Compute) Do(ctx context.Context, stmt []string) (models.Request, error) {
	if err := ctx.Err(); err != nil {
		return models.Request{}, err
	}

	cmd, key, value := c.p.Parse(stmt)

	cmdParsed, err := models.ParseCommand(cmd)
	if err != nil {
		return models.Request{}, err
	}

	req, err := models.NewRequest(cmdParsed, key, value)
	if err != nil {
		return models.Request{}, err
	}

	return req, nil
}
