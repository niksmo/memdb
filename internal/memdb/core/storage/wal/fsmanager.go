package wal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

const (
	minSegmentSize = 256
	fileNameLayout = "20060102150405_"
	fileExt        = ".json"
)

type JSONFileManager struct {
	logger         *slog.Logger
	dir            string
	maxSegmentSize int

	segmentSize int
	root        *os.Root      // segments root
	file        *os.File      // current segment
	bw          *bufio.Writer // wrap file for buffered write
}

func NewJSONFileManager(dir string, maxSegmentSize int) (*JSONFileManager, error) {
	if maxSegmentSize < minSegmentSize {
		return nil, fmt.Errorf("max segment size %d less than %d", maxSegmentSize, minSegmentSize)
	}

	m := &JSONFileManager{
		dir:            dir,
		maxSegmentSize: maxSegmentSize,
	}

	if err := m.initDir(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *JSONFileManager) WriteOps(operations []domain.Operation) error {
	const op = "JSONFileManager.WriteOps"

	if m.bw == nil {
		return errors.New("must call ReadEntries before write")
	}

	for _, operation := range operations {
		p, err := json.Marshal(operation)
		if err != nil {
			return fmt.Errorf("%s: marshal entry: %w", op, err)
		}

		p = append(p, '\n') // add new line for future decode stream

		requiredSize := len(p)
		freeSize := m.maxSegmentSize - m.segmentSize
		if requiredSize > freeSize {
			if err := m.applyNewSegment(); err != nil {
				return fmt.Errorf("%s: apply new segment: %w", op, err)
			}
		}

		n, err := m.bw.Write(p)
		if err != nil {
			return fmt.Errorf("%s: write %v: %w", op, operation, err)
		}

		m.segmentSize += n
	}

	if err := m.bw.Flush(); err != nil {
		return fmt.Errorf("%s: flush buffer %w", op, err)
	}

	if err := m.file.Sync(); err != nil {
		return fmt.Errorf("%s: sync file: %w", op, err)
	}

	return nil
}

func (m *JSONFileManager) ReadOps(ctx context.Context, fn func(domain.Operation) error) error {
	const op = "JSONFileManager.ReadOps"

	if m.root == nil {
		return errors.New("root is not initialized")
	}

	lastSegmentName, err := m.walkSegments(ctx, fn)

	if err != nil {
		return fmt.Errorf("%s: walk dir %w", op, err)
	}

	if lastSegmentName != "" {
		if err := m.applyExistingSegment(lastSegmentName); err != nil {
			return fmt.Errorf("%s: open last file for append %w", op, err)
		}
		return nil
	}

	if err = m.applyNewSegment(); err != nil {
		return fmt.Errorf("%s: apply new segment %w", op, err)
	}

	return nil
}

func (m *JSONFileManager) Close() error {
	const op = "JSONFileManager.Close"

	errs := append([]error(nil), m.closeSegment())

	if m.root != nil {
		errs = append(errs, m.root.Close())
		m.root = nil
	}

	err := errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (m *JSONFileManager) initDir() error {
	const op = "JSONFileManager.initDir"

	if m.dir == "" {
		return fmt.Errorf("%s: dir is empty", op)
	}

	rootPath := m.dir
	if !filepath.IsAbs(m.dir) { // relative to executable path
		binDir := filepath.Dir(os.Args[0])
		rootPath = filepath.Clean(filepath.Join(binDir, m.dir))
	}

	if err := os.MkdirAll(rootPath, 0750); err != nil {
		return fmt.Errorf("%s: make root dir %q: %w", op, m.dir, err)
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return fmt.Errorf("%s: open as root dir %q: %w", op, m.dir, err)
	}

	m.root = root

	return nil
}

func (m *JSONFileManager) walkSegments(
	ctx context.Context,
	fn func(domain.Operation) error,
) (lastSegmentName string, err error) {
	err = fs.WalkDir(m.root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err = ctx.Err(); err != nil {
			return context.Cause(ctx)
		}

		if path == "." || d.IsDir() {
			return nil
		}

		file, err := m.root.Open(path)
		if err != nil {
			return err
		}

		lastSegmentName = path

		if fn == nil {
			return nil
		}

		decoder := json.NewDecoder(file)

		for {
			var operation domain.Operation
			if err = decoder.Decode(&operation); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			if err = fn(operation); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return lastSegmentName, nil
}

func (m *JSONFileManager) closeSegment() error {
	if m.file == nil {
		return nil
	}

	if err := m.bw.Flush(); err != nil {
		return err
	}

	if err := m.file.Sync(); err != nil {
		return err
	}

	if err := m.file.Close(); err != nil {
		return err
	}

	m.file = nil
	m.bw = nil

	return nil
}

func (m *JSONFileManager) applyNewSegment() error {
	if m.file != nil {
		if err := m.closeSegment(); err != nil {
			return err
		}
	}

	name := generateSegmentName()

	file, err := m.root.Create(name)
	if err != nil {
		return err
	}

	m.file = file
	m.bw = bufio.NewWriter(file)
	m.segmentSize = 0

	return nil
}

func (m *JSONFileManager) applyExistingSegment(segmentName string) error {
	file, err := m.root.OpenFile(segmentName, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	m.file = file
	m.bw = bufio.NewWriter(file)

	return nil
}

func generateSegmentName() string {
	now := time.Now().UTC()
	nanoSecStr := strconv.Itoa(now.Nanosecond())
	return now.Format(fileNameLayout) + nanoSecStr + fileExt
}
