package pitch

import (
	"errors"
	"fmt"
	"io"
)

type Reader interface {
	Next() (*Header, error)
	Read([]byte) (int, error)
	Close() error
}

type internalReader interface {
	Reader
	reader() io.Reader
	discardContent() error
}

type reader struct {
	r             io.Reader
	contentReader io.LimitedReader
}

func NewReader(r io.Reader) Reader {
	return &reader{
		r: r,
		contentReader: io.LimitedReader{
			R: r,
		},
	}
}

func (rdr *reader) Next() (*Header, error) {
	if err := rdr.discardContent(); err != nil {
		return nil, fmt.Errorf("error discarding content: %w", err)
	}

	hdr, err := DecodeHeader(rdr.r)
	if err != nil {
		return nil, fmt.Errorf("error reading the next header: %w", err)
	}

	rdr.contentReader.N = int64(hdr.Size)

	return hdr, nil
}

func (rdr *reader) Read(b []byte) (int, error) {
	return rdr.contentReader.Read(b)
}

func (rdr *reader) discardContent() error {
	var (
		r = rdr.r
		n = rdr.contentReader.N
	)

	if n == 0 {
		return nil
	}

	if seeker, ok := r.(io.Seeker); ok {
		_, err := seeker.Seek(n, io.SeekCurrent)
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	_, err := io.CopyN(io.Discard, r, n)
	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
}

func (rdr *reader) reader() io.Reader {
	return rdr.r
}

func (r *reader) Close() error {
	var c, ok = r.r.(io.Closer)
	if ok {
		return c.Close()
	}

	return nil
}
