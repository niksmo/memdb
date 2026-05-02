package transport

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type reader struct {
	delim byte
	limit int
	rd    *bufio.Reader
}

func newReader(r io.Reader, limit int) *reader {
	return &reader{
		rd:    bufio.NewReader(r),
		delim: delim,
		limit: limit,
	}
}

func (r *reader) ReadBytes() ([]byte, error) {
	full, frag, n, err := r.collectFragments()
	if errors.Is(err, errMsgSize) {
		return nil, err
	}

	if n != 0 {
		n-- // cut delim
	}

	buf := make([]byte, n)
	n = 0
	for i := range full {
		n += copy(buf[n:], full[i])
	}
	copy(buf[n:], frag)
	return buf, err
}

func (r *reader) collectFragments() (fullBuffers [][]byte, finalFragment []byte, totalLen int, err error) {
	var frag []byte

	for {
		frag, err = r.rd.ReadSlice(r.delim)
		if err == nil { // got final fragment
			break
		}

		if !errors.Is(err, bufio.ErrBufferFull) { // unexpected error
			break
		}

		buf := bytes.Clone(frag)

		fullBuffers = append(fullBuffers, buf)
		totalLen += len(buf)
		if totalLen > r.limit {
			err = errMsgSize
			return
		}
	}

	totalLen += len(frag)
	if totalLen > r.limit {
		err = errMsgSize
	}

	return fullBuffers, frag, totalLen, err
}
