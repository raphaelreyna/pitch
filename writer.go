package pitch

import (
	"errors"
	"fmt"
	"io"
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

func (wtr *Writer) WriteHeader(name string, contentLength int64, data map[string][]string) (int, error) {
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

	h := Header{
		Name: name,
		Size: uint64(contentLength),
		Data: data,
	}
	payload := EncodeHeader(h)
	m, err = wtr.w.Write(payload)
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
