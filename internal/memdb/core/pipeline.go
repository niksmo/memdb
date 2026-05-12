package core

import (
	"context"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

type Compute interface {
	Do(ctx context.Context, payload []byte) (domain.Operation, error)
}

type Storage interface {
	Process(ctx context.Context, op domain.Operation) ([]byte, error)
}

type Pipeline struct {
	compute Compute
	storage Storage
}

func NewPipeline(c Compute, s Storage) *Pipeline {
	return &Pipeline{c, s}
}

func (p *Pipeline) Handle(ctx context.Context, payload []byte) ([]byte, error) {
	operation, err := p.compute.Do(ctx, payload)
	if err != nil {
		return nil, err
	}

	data, err := p.storage.Process(ctx, operation)
	if err != nil {
		return nil, err
	}

	return data, nil
}
