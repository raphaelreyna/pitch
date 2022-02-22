package pitch

import (
	"fmt"
	"io"
	"math"
)

type ByteRange struct {
	Start int64
	End   int64
}

type TableOfContents map[string]ByteRange

type TOCWriter struct {
	contentLength int64
	w             io.Writer
	toc           TableOfContents
	offset        int64
}

func NewTOCWriter(w io.Writer) *TOCWriter {
	return &TOCWriter{
		w:   w,
		toc: make(TableOfContents),
	}
}

func (wtr *TOCWriter) WriteHeader(name string, contentLength int64) error {
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

	n, err := w.Write([]byte(name))
	if err != nil {
		return err
	}

	wtr.offset += int64(n)

	if err := wtr.writeSize(contentLength); err != nil {
		return err
	}

	wtr.contentLength = contentLength
	wtr.toc[name] = ByteRange{
		Start: wtr.offset,
		End:   wtr.offset + contentLength,
	}

	return nil
}

func (wtr *TOCWriter) Write(b []byte) (int, error) {
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

	var m64 = int64(m)
	wtr.offset += m64
	wtr.contentLength -= m64

	if isTooLong {
		err = ErrWriteTooLong
	}

	return m, err
}

func (wtr *TOCWriter) Close() error {
	if wtr.w == nil {
		return ErrClosed
	}

	if err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	wtr.w = nil

	return nil
}

func (wtr *TOCWriter) pad() error {
	var (
		cl     = wtr.contentLength
		_, err = wtr.w.Write(make([]byte, cl))
	)

	wtr.offset += int64(cl)

	return err
}

func (wtr *TOCWriter) writeSize(size int64) error {
	if size < 1 {
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

	var m, err = wtr.w.Write(buf)
	if err != nil {
		return err
	}

	wtr.offset += int64(m)

	return err
}

func (wtr *TOCWriter) TableOfContents() TableOfContents {
	return wtr.toc
}
