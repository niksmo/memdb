package transport

import (
	"bufio"
	"io"
)

type writer struct {
	delim byte
	wr    *bufio.Writer
}

func newWriter(wr io.Writer) *writer {
	return &writer{
		wr:    bufio.NewWriter(wr),
		delim: delim,
	}
}

func (w *writer) Write(p []byte) (int, error) {
	n, err := w.wr.Write(p)
	if err != nil {
		return n, err
	}

	// according to the protocol
	if err := w.wr.WriteByte(w.delim); err != nil {
		return n, err
	}

	err = w.wr.Flush()
	return n, err
}
