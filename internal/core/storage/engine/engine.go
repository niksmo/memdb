package engine

import "errors"

var (
	ErrKeyNotFound = errors.New("not found")
)

type Engine struct {
	data map[string]string
}

func New() *Engine {
	return &Engine{
		data: make(map[string]string),
	}
}

func (e *Engine) Set(key string, value string) {
	e.data[key] = value
}

func (e *Engine) Get(key string) (string, error) {
	value, ok := e.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

func (e *Engine) Del(key string) {
	if _, ok := e.data[key]; ok {
		delete(e.data, key)
	}
}
