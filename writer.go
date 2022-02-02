package pitch

import (
	"errors"
	"fmt"
	"io"
	"math"
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

func (wtr *Writer) WriteHeader(name string, contentLength int64) error {
	var w = wtr.w
	if w == nil {
		return ErrClosed
	}

	if err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	var nameLength = int64(len(name))

	if err := wtr.writeSize(nameLength); err != nil {
		return err
	}

	if _, err := w.Write([]byte(name)); err != nil {
		return err
	}

	if err := wtr.writeSize(contentLength); err != nil {
		return err
	}

	wtr.contentLength = contentLength

	return nil
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

	if err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	wtr.w = nil

	return nil
}

func (wtr *Writer) pad() error {
	var (
		cl     = wtr.contentLength
		_, err = wtr.w.Write(make([]byte, cl))
	)

	return err
}

func (wtr *Writer) writeSize(size int64) error {
	if size < 0 {
		return ErrInvalidSize
	}

	var (
		n   = int(math.Floor(math.Log2(float64(size)))/7 + 1)
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

	var _, err = wtr.w.Write(buf)

	return err
}
