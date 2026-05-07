package wal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
)

func makeTestTempDir(t *testing.T, dirName string) string {
	t.Helper()

	initTempDir, err := os.MkdirTemp("", "wal_testing")
	if err != nil {
		t.Fatalf("os.MkdirTemp: %v", err)
	}

	t.Cleanup(func() {
		t.Helper()
		if err := os.RemoveAll(initTempDir); err != nil {
			t.Errorf("os.RemoveALL(initTempDir) = %v", err)
		}
	})

	return filepath.Join(initTempDir, dirName)
}

func newTestFSManager(t *testing.T, maxSegmentSize int) *JSONFileManager {
	t.Helper()

	dir := makeTestTempDir(t, "data/wal")

	m := &JSONFileManager{
		maxSegmentSize: maxSegmentSize,
		dir:            dir,
	}

	t.Cleanup(func() {
		t.Helper()
		if err := m.Close(); err != nil {
			t.Errorf("fsManager.Close() = %v", err)
		}
	})

	return m
}

func TestFSManager_InitDir(t *testing.T) {
	t.Parallel()

	t.Run("relative path = data/wal", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 1 << 10

		m := JSONFileManager{
			maxSegmentSize: maxSegmentSize,
			dir:            "data/wal",
		}

		err := m.initDir()
		require.NoError(t, err)
	})

	t.Run("absolute path = /data/wal", func(t *testing.T) {
		t.Parallel()

		initTempDir, err := os.MkdirTemp("", "wal_testing")
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			if err := os.RemoveAll(initTempDir); err != nil {
				t.Logf("os.RemoveALL(initTempDir) = %v", err)
			}
		})

		absoluteTempDir := filepath.Join(initTempDir, "data/wal")

		m := JSONFileManager{
			maxSegmentSize: 1 << 10,
			dir:            absoluteTempDir,
		}

		err = m.initDir()
		require.NoError(t, err)
	})
}

func TestFSManager_WriteReadEntries(t *testing.T) {
	t.Parallel()

	t.Run("read with cancelled context returns context.Canceled", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 1 << 10
		m := newTestFSManager(t, maxSegmentSize)

		wantEntriesCount := 0

		err := m.initDir()
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		var entriesCount int
		err = m.ReadOps(ctx, func(e domain.Operation) error {
			entriesCount++
			return nil
		})
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, wantEntriesCount, entriesCount)
		require.Nil(t, m.file)
	})

	t.Run("read with no segments returns zero and creates segment", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 1 << 10
		m := newTestFSManager(t, maxSegmentSize)

		wantEntriesCount := 0

		err := m.initDir()
		require.NoError(t, err)

		var entriesCount int
		err = m.ReadOps(t.Context(), func(e domain.Operation) error {
			entriesCount++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, wantEntriesCount, entriesCount)
		require.NotNil(t, m.file)
	})

	t.Run("write before read returns error", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 1 << 10
		m := newTestFSManager(t, maxSegmentSize)

		entries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey", Payload: "testValue"},
		}

		err := m.initDir()
		require.NoError(t, err)

		err = m.WriteOps(entries)
		require.Error(t, err)
	})

	t.Run("write after read persists data to file", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 1 << 10
		m := newTestFSManager(t, maxSegmentSize)

		entries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey", Payload: "testValue"},
		}

		wantReadCount := 0
		wantWrittenEntry := domain.Operation{Code: domain.OpSet, Key: "testKey", Payload: "testValue"}

		ctx := t.Context()

		err := m.initDir()
		require.NoError(t, err)

		var readCount int
		err = m.ReadOps(ctx, func(e domain.Operation) error {
			readCount++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, wantReadCount, readCount)

		err = m.WriteOps(entries)
		require.NoError(t, err)

		_, err = m.file.Seek(0, io.SeekStart)
		require.NoError(t, err)

		var writtenEntry domain.Operation
		err = json.NewDecoder(m.file).Decode(&writtenEntry)
		require.NoError(t, err)
		require.Equal(t, wantWrittenEntry, writtenEntry)
	})

	t.Run("second write persists in next segment", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 60 // each entry is 50 bytes + 1 byte is new line
		m := newTestFSManager(t, maxSegmentSize)

		entries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey1", Payload: "testValue1"},
			{Code: domain.OpSet, Key: "testKey2", Payload: "testValue2"},
		}

		wantSegmentsCount := 2
		wantEntries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey1", Payload: "testValue1"},
			{Code: domain.OpSet, Key: "testKey2", Payload: "testValue2"},
		}

		ctx := t.Context()

		err := m.initDir()
		require.NoError(t, err)

		err = m.ReadOps(ctx, nil)
		require.NoError(t, err)

		err = m.WriteOps(entries)
		require.NoError(t, err)

		var segmentsCount int
		entriesInSegments := make([]domain.Operation, 0)

		err = fs.WalkDir(m.root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
			require.NoError(t, err)

			if path == "." {
				return nil
			}

			file, err := m.root.Open(path)
			require.NoError(t, err)

			segmentsCount++

			decoder := json.NewDecoder(file)

			for {
				var e domain.Operation
				if err = decoder.Decode(&e); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					require.NoError(t, err)
				}
				entriesInSegments = append(entriesInSegments, e)
			}

			return nil
		})
		require.NoError(t, err)
		require.Equal(t, wantSegmentsCount, segmentsCount)
		require.Equal(t, wantEntries, entriesInSegments)
	})

	t.Run("end to end", func(t *testing.T) {
		t.Parallel()

		const maxSegmentSize = 120 // each entry is 50 bytes + 1 byte is new line
		m1 := newTestFSManager(t, maxSegmentSize)

		dir := m1.dir

		m1Entries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey1", Payload: "testValue1"},
			{Code: domain.OpSet, Key: "testKey2", Payload: "testValue2"},
			{Code: domain.OpSet, Key: "testKey3", Payload: "testValue3"},
		}

		ctx := t.Context()

		wantRestoredEntries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey1", Payload: "testValue1"},
			{Code: domain.OpSet, Key: "testKey2", Payload: "testValue2"},
			{Code: domain.OpSet, Key: "testKey3", Payload: "testValue3"},
		}

		wantSegmentsCount := 2

		wantEntries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey1", Payload: "testValue1"},
			{Code: domain.OpSet, Key: "testKey2", Payload: "testValue2"},
			{Code: domain.OpSet, Key: "testKey3", Payload: "testValue3"},
			{Code: domain.OpSet, Key: "testKey4", Payload: "testValue4"},
		}

		err := m1.initDir()
		require.NoError(t, err)

		err = m1.ReadOps(ctx, nil)
		require.NoError(t, err)

		err = m1.WriteOps(m1Entries)
		require.NoError(t, err)

		err = m1.Close() // gracefully shutdown
		require.NoError(t, err)

		// next process bootstrap and restore
		m2 := JSONFileManager{
			maxSegmentSize: maxSegmentSize,
			dir:            dir,
		}

		err = m2.initDir()
		require.NoError(t, err)

		restoredEntries := make([]domain.Operation, 0)
		err = m2.ReadOps(ctx, func(e domain.Operation) error {
			restoredEntries = append(restoredEntries, e)
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, wantRestoredEntries, restoredEntries)

		m2Entries := []domain.Operation{
			{Code: domain.OpSet, Key: "testKey4", Payload: "testValue4"},
		}

		err = m2.WriteOps(m2Entries)
		require.NoError(t, err)

		var segmentsCount int
		entriesInSegments := make([]domain.Operation, 0)

		err = fs.WalkDir(m2.root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
			require.NoError(t, err)

			if path == "." {
				return nil
			}

			file, err := m2.root.Open(path)
			require.NoError(t, err)

			segmentsCount++

			decoder := json.NewDecoder(file)

			for {
				var e domain.Operation
				if err = decoder.Decode(&e); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					require.NoError(t, err)
				}
				entriesInSegments = append(entriesInSegments, e)
			}

			return nil
		})
		require.NoError(t, err)
		require.Equal(t, wantSegmentsCount, segmentsCount)
		require.Equal(t, wantEntries, entriesInSegments)
	})
}
