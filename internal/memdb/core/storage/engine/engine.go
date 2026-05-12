package engine

import (
	"errors"
	"sync"
)

var (
	NotFound = errors.New("not found")
)

type Engine struct {
	mu    sync.RWMutex
	vault map[string]string
}

func New() *Engine {
	return &Engine{
		vault: make(map[string]string),
	}
}

func (e *Engine) Set(key string, payload string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vault[key] = payload
}

func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock()
	data, ok := e.vault[key]
	e.mu.RUnlock()

	if !ok {
		return "", NotFound
	}
	return data, nil
}

func (e *Engine) Del(key string) {
	e.mu.Lock()
	_, ok := e.vault[key]
	e.mu.Unlock()

	if ok {
		delete(e.vault, key)
	}
}
