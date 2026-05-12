package wal

import (
	"context"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/internal/memdb/core/domain"
	"github.com/niksmo/memdb/pkg/logger"
)

type mockFileManager struct {
	mu               sync.Mutex
	capturedWriteOps [][]domain.Operation
	writeOpsRet      error

	readOpsRet error

	closeCalled bool
	closeRet    error
}

func (m *mockFileManager) CapturedWriteOps() [][]domain.Operation {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([][]domain.Operation, len(m.capturedWriteOps))
	copy(result, m.capturedWriteOps)
	return result
}

func (m *mockFileManager) WriteOps(operations []domain.Operation) error {
	m.mu.Lock()
	m.capturedWriteOps = append(m.capturedWriteOps, operations)
	m.mu.Unlock()
	return m.writeOpsRet
}

func (m *mockFileManager) ReadOps(context.Context, func(domain.Operation) error) error {
	return m.readOpsRet
}

func (m *mockFileManager) Close() error {
	m.closeCalled = true
	return m.closeRet
}

func withLock(mu *sync.Mutex, fn func()) {
	mu.Lock()
	fn()
	mu.Unlock()
}

func TestImplementation(t *testing.T) {
	t.Parallel()

	t.Run("flush on tick", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			const (
				first = iota
				second
			)

			log, _ := logger.NewObservable("info")

			mockFM := &mockFileManager{
				writeOpsRet: nil,
				closeRet:    nil,
			}

			options := Options{
				Logger:       log,
				FileManager:  mockFM,
				BatchSize:    4,
				FlushingTick: 100 * time.Millisecond,
			}

			ctx := context.Background()

			imp, err := New(options)
			require.NoError(t, err)

			go func() {
				rErr := imp.Run()
				require.NoError(t, rErr)
			}()
			synctest.Wait()

			mu := new(sync.Mutex)
			results := make([]bool, 2)

			go func() { // must be in first flush
				err := imp.WriteLog(ctx, domain.OpSet, "key1", "val1")
				require.NoError(t, err)
				withLock(mu, func() {
					results[first] = true
				})
			}()

			go func() { // must be in second flush
				time.Sleep(120 * time.Millisecond)
				err := imp.WriteLog(ctx, domain.OpSet, "key2", "val2")
				require.NoError(t, err)
				withLock(mu, func() {
					results[second] = true
				})
			}()
			synctest.Wait()

			// not flushed yet
			require.Len(t, mockFM.CapturedWriteOps(), 0)
			withLock(mu, func() {
				require.Equal(t, []bool{first: false, second: false}, results)
			})

			// first flush
			time.Sleep(101 * time.Millisecond)
			require.Len(t, mockFM.CapturedWriteOps(), 1)
			withLock(mu, func() {
				require.Equal(t, []bool{first: true, second: false}, results)
			})

			// second steel wait
			time.Sleep(21 * time.Millisecond)
			require.Len(t, mockFM.CapturedWriteOps(), 1)
			withLock(mu, func() {
				require.Equal(t, []bool{first: true, second: false}, results)
			})

			// second flush
			time.Sleep(81 * time.Millisecond)
			require.Len(t, mockFM.CapturedWriteOps(), 2)
			withLock(mu, func() {
				require.Equal(t, []bool{first: true, second: true}, results)
			})

			err = imp.Close()
			require.NoError(t, err)
			synctest.Wait()
		})
	})

	t.Run("flush on batch limit exceeded", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			const (
				first = iota
				second
			)

			log, _ := logger.NewObservable("info")

			mockFM := &mockFileManager{
				writeOpsRet: nil,
				closeRet:    nil,
			}

			options := Options{
				Logger:       log,
				FileManager:  mockFM,
				BatchSize:    2,
				FlushingTick: 100 * time.Millisecond,
			}

			ctx := context.Background()

			imp, err := New(options)
			require.NoError(t, err)

			go func() {
				rErr := imp.Run()
				require.NoError(t, rErr)
			}()
			synctest.Wait()

			mu := new(sync.Mutex)
			results := make([]bool, 2)

			go func() {
				err := imp.WriteLog(ctx, domain.OpSet, "key1", "val1")
				require.NoError(t, err)
				withLock(mu, func() {
					results[first] = true
				})
			}()

			go func() {
				time.Sleep(50 * time.Millisecond)
				err := imp.WriteLog(ctx, domain.OpSet, "key2", "val2")
				require.NoError(t, err)
				withLock(mu, func() {
					results[second] = true
				})
			}()
			synctest.Wait()

			// not flushed yet
			require.Len(t, mockFM.CapturedWriteOps(), 0)
			withLock(mu, func() {
				require.Equal(t, []bool{first: false, second: false}, results)
			})

			// flush
			time.Sleep(51 * time.Millisecond) // <- flush timeout is 100ms
			require.Len(t, mockFM.CapturedWriteOps(), 1)
			withLock(mu, func() {
				require.Equal(t, []bool{first: true, second: true}, results)
			})

			err = imp.Close()
			require.NoError(t, err)
			synctest.Wait()
		})
	})

	t.Run("flush on close", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			const (
				first = iota
				second
			)

			log, _ := logger.NewObservable("info")

			mockFM := &mockFileManager{
				writeOpsRet: nil,
				closeRet:    nil,
			}

			options := Options{
				Logger:       log,
				FileManager:  mockFM,
				BatchSize:    4,
				FlushingTick: 100 * time.Millisecond,
			}

			ctx := context.Background()

			imp, err := New(options)
			require.NoError(t, err)

			go func() {
				rErr := imp.Run()
				require.NoError(t, rErr)
			}()
			synctest.Wait()

			mu := new(sync.Mutex)
			results := make([]bool, 2)

			go func() {
				err := imp.WriteLog(ctx, domain.OpSet, "key1", "val1")
				require.NoError(t, err)
				withLock(mu, func() {
					results[first] = true
				})
			}()

			go func() {
				time.Sleep(50 * time.Millisecond)
				err := imp.WriteLog(ctx, domain.OpSet, "key2", "val2")
				require.NoError(t, err)
				withLock(mu, func() {
					results[second] = true
				})
			}()
			synctest.Wait()

			time.Sleep(51 * time.Millisecond)

			// not flushed yet
			require.Len(t, mockFM.CapturedWriteOps(), 0)
			withLock(mu, func() {
				require.Equal(t, []bool{first: false, second: false}, results)
			})

			err = imp.Close()
			require.NoError(t, err)
		})
	})

	t.Run("canceled write log", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			log, _ := logger.NewObservable("info")

			mockFM := &mockFileManager{
				writeOpsRet: nil,
				closeRet:    nil,
			}

			options := Options{
				Logger:       log,
				FileManager:  mockFM,
				BatchSize:    4,
				FlushingTick: 100 * time.Millisecond,
			}

			ctx, cancel := context.WithCancel(context.Background())

			imp, err := New(options)
			require.NoError(t, err)

			go func() {
				rErr := imp.Run()
				require.NoError(t, rErr)
			}()
			synctest.Wait()

			cancel()

			go func() { // must be in first flush
				err := imp.WriteLog(ctx, domain.OpSet, "key1", "val1")
				require.ErrorIs(t, err, context.Canceled)
			}()
			synctest.Wait()

			err = imp.Close()
			require.NoError(t, err)
		})
	})
}
