package pitch

import (
	"errors"
	"io"
)

type catReader struct {
	r             Reader
	readers       []Reader
	contentReader io.LimitedReader
	offset        int64
}

func (mr *catReader) Next() (*Header, error) {
	if mr.r == nil {
		return nil, io.EOF
	}

	r, ok := mr.r.(internalReader)
	if !ok {
		return nil, errors.New("unrecognized reader")
	}

	hdr, e := r.Next()
	if errors.Is(e, io.EOF) {
		if len(mr.readers) == 0 {
			return nil, io.EOF
		}
		mr.r = mr.readers[0]
		mr.readers = mr.readers[1:]
		return mr.Next()
	}

	return hdr, e
}

func (mr *catReader) Read(b []byte) (int, error) {
	if mr.r == nil {
		return 0, io.EOF
	}

	r, ok := mr.r.(internalReader)
	if !ok {
		return 0, errors.New("unrecognized reader")
	}

	n, e := r.Read(b)
	if errors.Is(e, io.EOF) {
		if len(mr.readers) == 0 {
			return 0, io.EOF
		}
		mr.r = mr.readers[0]
		mr.readers = mr.readers[1:]
		return mr.Read(b)
	}

	return n, e
}

func (mr *catReader) Close() error {
	var err error
	for _, r := range mr.readers {
		if e := r.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func (mr *catReader) discardContent() error {
	if mr.r == nil {
		return nil
	}

	r, ok := mr.r.(internalReader)
	if !ok {
		return errors.New("unrecognized reader")
	}

	e := r.discardContent()
	if errors.Is(e, io.EOF) {
		if len(mr.readers) == 0 {
			return nil
		}
		mr.r = mr.readers[0]
		mr.readers = mr.readers[1:]
		return mr.discardContent()
	}

	return e
}

func (mr *catReader) reader() io.Reader {
	if r, ok := mr.r.(internalReader); ok {
		return r.reader()
	}
	return mr.contentReader.R
}

func Cat(r ...Reader) Reader {
	var (
		readers = make([]Reader, len(r))
		offset  int64
	)
	copy(readers, r)
	return &catReader{
		r:       readers[0],
		readers: readers[1:],
		offset:  offset,
	}
}
