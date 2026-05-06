package engine

import (
	"errors"
	"sync"
)

var (
	NotFound = errors.New("not found")
)

type Engine struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *Engine {
	return &Engine{
		data: make(map[string]string),
	}
}

func (e *Engine) Set(key string, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data[key] = value
}

func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock()
	value, ok := e.data[key]
	e.mu.RUnlock()

	if !ok {
		return "", NotFound
	}
	return value, nil
}

func (e *Engine) Del(key string) {
	e.mu.Lock()
	_, ok := e.data[key]
	e.mu.Unlock()

	if ok {
		delete(e.data, key)
	}
}
