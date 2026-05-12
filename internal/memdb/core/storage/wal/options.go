package wal

import (
	"errors"
	"fmt"
	"log/slog"
	"time"
)

const (
	minBatchSize    = 1
	minFlushingTick = time.Millisecond
)

type Options struct {
	Logger       *slog.Logger
	BatchSize    int
	FlushingTick time.Duration
	FileManager  FileManager
}

func (o *Options) validate() error {
	if o.Logger == nil {
		return errors.New("logger is not set")
	}

	if o.BatchSize < minBatchSize {
		return fmt.Errorf("batch size %d less than %d", o.BatchSize, minBatchSize)
	}

	if o.FlushingTick < 0 {
		return fmt.Errorf("flushing tick %d less than %d", o.FlushingTick, minFlushingTick)
	}

	if o.FileManager == nil {
		return errors.New("file manager is not set")
	}

	return nil
}
