package pitch

import (
	"errors"
	"fmt"
	"io"
)

type Reader struct {
	r             io.Reader
	contentReader io.LimitedReader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
		contentReader: io.LimitedReader{
			R: r,
		},
	}
}

func (rdr *Reader) Next() (name string, err error) {
	if err := rdr.discardContent(); err != nil {
		return "", fmt.Errorf("error reading the next header: %w", err)
	}

	size, err := rdr.readSize()
	if errors.Is(err, io.EOF) {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("error reading file name size: %w", err)
	}

	var (
		r       = rdr.r
		nameBuf = make([]byte, size)
	)

	_, err = io.ReadFull(r, nameBuf)
	if err != nil {
		return "", fmt.Errorf("error reading header byte: %w", err)
	}

	size, err = rdr.readSize()
	if err != nil {
		return "", fmt.Errorf("error reading file size: %w", err)
	}

	rdr.contentReader.N = size

	return string(nameBuf), nil
}

func (rdr *Reader) Read(b []byte) (int, error) {
	return rdr.contentReader.Read(b)
}

func (rdr *Reader) discardContent() error {
	var (
		r = rdr.r
		n = rdr.contentReader.N
	)

	_, err := io.CopyN(io.Discard, r, n)
	if errors.Is(err, io.EOF) {
		return nil
	}

	return err
}

func (rdr *Reader) readSize() (int64, error) {
	const lbMask = 0b00000001
	var (
		r    = rdr.r
		size int64
		buf  = make([]byte, 1)
	)

	for {
		_, err := r.Read(buf)
		if err != nil {
			return 0, err
		}

		var (
			x   = buf[0]
			end = x&lbMask == lbMask
		)

		size <<= 8
		size |= int64(x)
		size >>= 1

		if end {
			break
		}
	}

	return size, nil
}
