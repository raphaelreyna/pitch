package pitch

import (
	"errors"
	"fmt"
	"io"
	"math"

	"golang.org/x/exp/constraints"
)

var (
	ErrWriteTooLong = errors.New("pitch: write too long")
	ErrClosed       = errors.New("pitch: writer is closed")
	ErrInvalidSize  = errors.New("pitch: invalid size")
)

type Writer struct {
	contentLength int64
	w             io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

func (wtr *Writer) WriteHeader(name string, contentLength int64) (int, error) {
	var (
		w = wtr.w
		n int
	)

	if w == nil {
		return 0, ErrClosed
	}

	if contentLength < 0 {
		return 0, ErrInvalidSize
	}

	m, err := wtr.pad()
	n += m
	if err != nil {
		return n, fmt.Errorf("error padding file: %w", err)
	}

	var nameLength = int64(len(name))

	m, err = wtr.writeSize(nameLength)
	n += m
	if err != nil {
		return n, err
	}

	m, err = w.Write([]byte(name))
	n += m
	if err != nil {
		return n, err
	}

	m, err = wtr.writeSize(contentLength)
	n += m
	if err != nil {
		return n, err
	}

	wtr.contentLength = contentLength

	return n, nil
}

func (wtr *Writer) Write(b []byte) (int, error) {
	var w = wtr.w
	if w == nil {
		return 0, ErrClosed
	}

	var (
		n  = int64(len(b))
		cl = wtr.contentLength
	)

	if cl == 0 {
		return 0, ErrWriteTooLong
	}

	var isTooLong bool
	if cl = wtr.contentLength; cl < n {
		n = cl
		isTooLong = true
	}

	m, err := w.Write(b[:n])
	if err != nil {
		return m, err
	}

	wtr.contentLength -= int64(m)

	if isTooLong {
		err = ErrWriteTooLong
	}

	return m, err
}

func (wtr *Writer) Close() error {
	if wtr.w == nil {
		return ErrClosed
	}

	if _, err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	wtr.w = nil

	return nil
}

func (wtr *Writer) pad() (int, error) {
	return wtr.w.Write(make([]byte, wtr.contentLength))
}

func (wtr *Writer) writeSize(size int64) (int, error) {
	if size < 0 {
		return 0, ErrInvalidSize
	}

	var (
		n   = sizeWriteSize(size)
		buf = make([]byte, n)
	)

	for idx := n - 1; -1 < idx; idx-- {
		// grab the lowest 7 bits
		var b byte = byte(size & 0b01111111)
		size >>= 7

		// shift to the right to free the low bit
		b <<= 1

		// set the low bit high if we're done encoding
		if idx == n-1 {
			b |= 1
		}

		buf[idx] = b
	}

	return wtr.w.Write(buf)
}

func sizeWriteSize[N constraints.Integer](size N) int {
	return int(math.Floor(math.Log2(float64(size)))/7 + 1)
}
