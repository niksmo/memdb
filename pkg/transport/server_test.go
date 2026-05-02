package transport

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/pkg/logger"
)

/*
#######################################    MOCKs    #######################################
*/

// mockHandler implements Handler interface and used in tests.
type mockHandler struct {
	res     []byte
	err     error
	panicFn func()

	callCount int
	capturedP []byte
}

func (m *mockHandler) Handle(ctx context.Context, p []byte) ([]byte, error) {
	m.callCount++
	m.capturedP = p

	if m.panicFn != nil {
		m.panicFn()
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return m.res, m.err
}

// mockTCPListener implements net.Listener interface for testing in synctest.Test.
type mockTCPListener struct {
	net.Listener

	conns  []net.Conn
	connCh chan net.Conn
}

func (l *mockTCPListener) Accept() (net.Conn, error) {
	c, ok := <-l.connCh
	if !ok {
		return nil, net.ErrClosed
	}
	return c, nil
}

func (l *mockTCPListener) Close() error {
	for i := range l.conns { // close all opened connections
		_ = l.conns[i].Close()
	}

	close(l.connCh)

	return nil
}

func (l *mockTCPListener) establishConn() (clientConn net.Conn) {
	srvConn, clConn := net.Pipe()

	l.conns = append(l.conns, srvConn) // store connections for close on Close method.

	go func() {
		l.connCh <- srvConn
	}()

	return clConn
}

/*
###################################    CONSTRUCTORs    ####################################
*/

// newTestLogger create new slog.Logger instance which write logs in observer.
func newTestLogger(t *testing.T, level string) (l *slog.Logger, observer *strings.Builder) {
	t.Helper()

	observer = new(strings.Builder)

	return logger.New(observer, level), observer
}

// newTestConfig create Config with more frequency settings.
func newTestConfig() Config {
	return Config{
		MaxMessageSize: 4 << 10,
		IdleTimeout:    time.Minute,
		Address:        "127.0.0.1:",
		MaxConnections: 2,
	}
}

// newMockListener create new mockTCPListener instance  with buffered connection channel.
func newMockListener(buffer int) *mockTCPListener {
	return &mockTCPListener{
		connCh: make(chan net.Conn, buffer),
	}

}

// newTestServer create new Server instance.
// The instance leverages a native TCP listener for its underlying transport layer.
func newTestServer(t *testing.T, l *slog.Logger, e Handler, c Config) (srv *Server) {
	t.Helper()

	srv, err := NewListener(l, e, c)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, srv.Close())
	})

	return srv
}

// newTestClient create new nature TCP connection by net.Dialer.
func newTestClient(t *testing.T, addr string) net.Conn {
	t.Helper()

	var d net.Dialer
	conn, err := d.DialContext(t.Context(), "tcp", addr)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, conn.Close())
	})

	return conn
}

/*
#######################################    TESTs    #######################################
*/

func TestServer_Listen_ReadResponse(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	req := []byte("request\n")

	res := []byte("response")

	l, _ := newTestLogger(t, "debug")

	handler := &mockHandler{
		res: res,
	}

	config := newTestConfig()

	srv := newTestServer(t, l, handler, config)
	addr := srv.lis.Addr().String()

	expectedCallCount := 1
	expectedCaptured := []byte("request")
	expectedResp := []byte("response\n")

	go func() {
		_ = srv.Listen(ctx)
	}()

	cl := newTestClient(t, addr)

	_, err := cl.Write(req)
	require.NoError(t, err)

	buf := make([]byte, 1<<10)
	n, err := cl.Read(buf)
	require.NoError(t, err)
	require.Equal(t, expectedResp, buf[:n])

	require.Equal(t, expectedCallCount, handler.callCount)
	require.Equal(t, expectedCaptured, handler.capturedP)
}

func TestServer_Listen_ExecuterError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	req := []byte("request\n")

	l, _ := newTestLogger(t, "debug")

	handlerErr := errors.New("test error")

	handler := &mockHandler{
		err: handlerErr,
	}

	config := newTestConfig()

	srv := newTestServer(t, l, handler, config)
	addr := srv.lis.Addr().String()

	expectedCallCount := 1
	expectedCaptured := []byte("request")
	expectedResponse := []byte(handlerErr.Error() + "\n")

	go func() {
		_ = srv.Listen(ctx)
	}()

	cl := newTestClient(t, addr)

	_, err := cl.Write(req)
	require.NoError(t, err)

	buf := make([]byte, 1<<10)
	n, err := cl.Read(buf)
	require.NoError(t, err)
	require.Equal(t, expectedResponse, buf[:n])

	require.Equal(t, expectedCallCount, handler.callCount)
	require.Equal(t, expectedCaptured, handler.capturedP)
}

func TestServer_Listen_ExecuterPanic(t *testing.T) {
	t.Parallel()

	l, observer := newTestLogger(t, "debug")

	handler := &mockHandler{
		panicFn: func() {
			panic("boom")
		},
	}

	config := newTestConfig()

	synctest.Test(t, func(t *testing.T) {
		mockLis := newMockListener(0)

		srv := Server{
			logger:  l,
			handler: handler,
			lis:     mockLis,
			cfg:     config,
		}

		t.Cleanup(func() {
			_ = srv.Close() // mockLis returning nil
		})

		ctx := t.Context()

		go func() {
			require.NotPanics(t, func() {
				_ = srv.Listen(ctx)
			})
		}()

		synctest.Wait() // wait for Listen is started

		cliConn := mockLis.establishConn()

		synctest.Wait() // wait for accept connection

		require.Equal(t, 1, int(srv.connCount.Load()), "expect one connection")

		n, err := cliConn.Write([]byte("request\n"))
		require.NoError(t, err)
		require.NotZero(t, n)

		synctest.Wait() // wait for handler exec and panic recovered but close connection

		log := observer.String()
		require.Contains(t, log, "connection closed")
		require.Contains(t, log, "panic recovered")

		require.Zero(t, srv.connCount.Load(), "connection should is closed")

		buf := make([]byte, 4<<10)
		n, err = cliConn.Read(buf)
		require.ErrorIs(t, err, io.EOF)
		require.Zero(t, n)
	})
}

func TestServer_Listen_ContextCancelled(t *testing.T) {
	t.Parallel()

	l, _ := newTestLogger(t, "debug")
	handler := &mockHandler{}
	config := newTestConfig()

	synctest.Test(t, func(t *testing.T) {
		lis := newMockListener(0)

		srv := Server{
			logger:  l,
			handler: handler,
			lis:     lis,
			cfg:     config,
		}

		ctx, cancel := context.WithCancel(t.Context())

		go func() {
			err := srv.Listen(ctx)
			require.ErrorIs(t, err, context.Canceled)

			err = srv.Close()
			require.NoError(t, err)
		}()

		synctest.Wait() // wait for starting accept loop

		clConn := lis.establishConn() // establish connection

		synctest.Wait() // waiting for the client read

		cancel() // interrupt memdb listening

		synctest.Wait() // waiting for the memdb to close

		// the client is trying to write,
		n, err := clConn.Write([]byte("hello world\n"))
		require.ErrorIs(t, err, io.ErrClosedPipe)
		require.Zero(t, n)
	})
}

func TestServer_Listen_ConnLimitExceeded(t *testing.T) {
	t.Parallel()

	l, _ := newTestLogger(t, "debug")

	handler := &mockHandler{}

	config := newTestConfig()
	config.MaxConnections = 1 // only one connection

	synctest.Test(t, func(t *testing.T) {
		connBuffer := 2
		mockLis := newMockListener(connBuffer)

		srv := Server{
			logger:  l,
			handler: handler,
			lis:     mockLis,
			cfg:     config,
		}
		t.Cleanup(func() {
			_ = srv.Close() // mockList returning nil
		})

		ctx := t.Context()

		go func() {
			_ = srv.Listen(ctx)
		}()

		synctest.Wait() // wait for Listen is started

		clConn1 := mockLis.establishConn()
		_ = clConn1

		synctest.Wait() // wait for accept first

		clConn2 := mockLis.establishConn() // second connection

		synctest.Wait() // wait for accept and then close second connection

		n, err := clConn2.Write([]byte("hello world\n"))
		require.ErrorIs(t, err, io.ErrClosedPipe)
		require.Zero(t, n)
	})
}

func TestServer_Listen_MsgSizeMoreThenLimit(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	req := []byte("request\n") // 8 bytes

	l, _ := newTestLogger(t, "debug")

	handler := &mockHandler{}

	config := newTestConfig()
	config.MaxMessageSize = 4 // 4 bytes is max size

	srv := newTestServer(t, l, handler, config)
	addr := srv.lis.Addr().String()

	expectedCallCount := 0
	expectedErrMsgSubstr := "data size exceeds limit"

	go func() {
		_ = srv.Listen(ctx)
	}()

	cl := newTestClient(t, addr)

	_, err := cl.Write(req)
	require.NoError(t, err)

	buf := make([]byte, 4<<10)
	n, err := cl.Read(buf) // receive error message
	require.NoError(t, err)
	require.Contains(t, string(buf[:n]), expectedErrMsgSubstr)

	n, err = cl.Read(buf[0:]) // connection closed
	require.ErrorIs(t, err, io.EOF)
	require.Zero(t, n)

	require.Equal(t, expectedCallCount, handler.callCount)
}

func TestServer_Listen_ClientReadTimeoutExceeded(t *testing.T) {
	t.Parallel()

	l, _ := newTestLogger(t, "debug")
	handler := &mockHandler{}
	config := newTestConfig()

	synctest.Test(t, func(t *testing.T) {
		mockLis := newMockListener(0)

		srv := Server{
			logger:  l,
			handler: handler,
			cfg:     config,
			lis:     mockLis,
		}
		t.Cleanup(func() {
			_ = srv.Close() // mock returning nil
		})

		ctx := t.Context()

		go func() {
			_ = srv.Listen(ctx)
		}()

		synctest.Wait() // listener listen

		cliConn := mockLis.establishConn()

		synctest.Wait() // listener accept conn

		require.Equal(t, int32(1), srv.connCount.Load(), "should have one connection")

		_, err := cliConn.Write([]byte("request\n"))
		require.NoError(t, err)

		time.Sleep(config.IdleTimeout) // client don't read

		synctest.Wait() // listener close conn

		require.Equal(t, int32(0), srv.connCount.Load(), "should have zero connections")

		// client wake up and try read
		buf := make([]byte, 4<<10)
		_, err = cliConn.Read(buf)
		require.ErrorIs(t, err, io.EOF, "client should receive EOF error")
	})
}

func TestServer_Listen_ClientFirstWriteTimeoutExceeded(t *testing.T) {
	t.Parallel()

	l, _ := newTestLogger(t, "debug")

	handler := &mockHandler{
		res: []byte("response"),
	}

	config := newTestConfig()

	synctest.Test(t, func(t *testing.T) {
		mockLis := newMockListener(0)

		srv := Server{
			logger:  l,
			handler: handler,
			cfg:     config,
			lis:     mockLis,
		}
		t.Cleanup(func() {
			_ = srv.Close() // mockLis returning nil
		})

		ctx := t.Context()

		go func() {
			_ = srv.Listen(ctx)
		}()

		synctest.Wait() // listener listen

		cliConn := mockLis.establishConn()

		synctest.Wait() // listener accept conn

		require.Equal(t, int32(1), srv.connCount.Load(), "should have one connection")

		time.Sleep(config.IdleTimeout) // the client does not write

		synctest.Wait() // listener close conn

		require.Equal(t, int32(0), srv.connCount.Load(), "should have zero connections")

		// client tries to write, but it's too late
		_, err := cliConn.Write([]byte("request\n"))
		require.Error(t, err)
	})
}

func TestServer_Listen_ClientSecondWriteTimeoutExceeded(t *testing.T) {
	t.Parallel()

	l, _ := newTestLogger(t, "debug")

	handler := &mockHandler{
		res: []byte("response"),
	}

	config := newTestConfig()

	synctest.Test(t, func(t *testing.T) {
		mockLis := newMockListener(0)

		srv := Server{
			logger:  l,
			handler: handler,
			cfg:     config,
			lis:     mockLis,
		}
		t.Cleanup(func() {
			_ = srv.Close() // mockLis returning nil
		})

		ctx := t.Context()

		go func() {
			_ = srv.Listen(ctx)
		}()

		synctest.Wait() // listener listen

		cliConn := mockLis.establishConn()

		synctest.Wait() // listener accept conn

		require.Equal(t, int32(1), srv.connCount.Load(), "should have one connection")

		_, err := cliConn.Write([]byte("request_1\n")) // client send first request
		require.NoError(t, err)

		// client read response
		buf := make([]byte, 4<<10)
		_, err = cliConn.Read(buf)
		require.NoError(t, err)

		time.Sleep(config.IdleTimeout) // the client does not write a second time

		synctest.Wait() // listener close conn

		require.Equal(t, int32(0), srv.connCount.Load(), "should have zero connections")

		// client tries to write, but it's too late
		_, err = cliConn.Write([]byte("request_2\n"))
		require.Error(t, err)
	})
}
