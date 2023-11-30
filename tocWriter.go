package pitch

import (
	"fmt"
	"io"
)

type ByteRange struct {
	Start int64
	End   int64
}

type HeaderItem struct {
	Size  uint64
	Data  map[string][]string
	Start int64
	End   int64
}

type TableOfContents map[string]*HeaderItem

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

func (wtr *TOCWriter) WriteHeader(name string, contentLength int64, data map[string][]string) error {
	var w = wtr.w
	if w == nil {
		return ErrClosed
	}
	if contentLength < 1 {
		return ErrInvalidSize
	}

	if err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	hdr := Header{
		Name: name,
		Size: uint64(contentLength),
		Data: data,
	}
	payload := EncodeHeader(hdr)
	wtr.offset += int64(len(payload))
	_, err := w.Write(payload)
	if err != nil {
		return err
	}

	wtr.contentLength = contentLength
	wtr.toc[name] = &HeaderItem{
		Size:  uint64(contentLength),
		Data:  nil,
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

func (wtr *TOCWriter) TableOfContents() TableOfContents {
	return wtr.toc
}
