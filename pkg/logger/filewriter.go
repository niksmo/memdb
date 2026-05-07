package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileWriter struct {
	file *os.File
}

func NewFileWriter(output string) (*FileWriter, error) {
	dir := filepath.Dir(output)
	if !filepath.IsAbs(dir) {
		return nil, errors.New("log output destination must be defined using absolute path")
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("create log output path: %v", err)
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return nil, fmt.Errorf("open log output path: %v", err)
	}

	return &FileWriter{file}, nil
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	return fw.file.Write(p)
}

func (fw *FileWriter) Close() error {
	return fw.file.Close()
}
