package pitch

import (
	"bytes"
	"crypto/rand"
	"io"
)

type inMemoryReader struct {
	buf *bytes.Buffer
	r   *reader
}

func NewContents(hdr ...*Header) map[*Header][]byte {
	data := make(map[*Header][]byte, len(hdr))
	for _, h := range hdr {
		if h.Size == 0 {
			h.Size = 32
		}
		data[h] = make([]byte, h.Size)
		_, err := rand.Read(data[h])
		if err != nil {
			panic(err)
		}
	}
	return data
}

func NewReaderFromMap(contents map[*Header][]byte) (Reader, error) {
	var (
		buf = bytes.NewBuffer(nil)
		w   = NewWriter(buf)
	)
	for hdr, content := range contents {
		_, err := w.WriteHeader(hdr.Name, int64(hdr.Size), hdr.Data)
		if err != nil {
			return nil, err
		}
		_, err = w.Write(content)
		if err != nil {
			return nil, err
		}
	}
	err := w.Close()
	if err != nil {
		return nil, err
	}

	return &inMemoryReader{
		buf: buf,
		r: &reader{
			r: buf,
			contentReader: io.LimitedReader{
				R: buf,
			},
		},
	}, nil
}

func (mr *inMemoryReader) Close() error {
	return nil
}
func (mr *inMemoryReader) Next() (*Header, error) {
	return mr.r.Next()
}

func (mr *inMemoryReader) Read(b []byte) (int, error) {
	return mr.r.Read(b)
}

func (mr *inMemoryReader) reader() io.Reader {
	return mr.r.reader()
}

func (mr *inMemoryReader) discardContent() error {
	return mr.r.discardContent()
}
