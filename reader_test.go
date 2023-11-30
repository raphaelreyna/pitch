package pitch

import (
	"errors"
	"io"
	"testing"

	"github.com/matryer/is"
)

func TestReader_ReadSize_DiscardErr(t *testing.T) {
	var (
		is = is.New(t)
		er = errReader{
			errors: []error{
				errors.New("ERROR"),
			},
		}
		r      = NewReader(&er)
		rr, ok = r.(*reader)
	)
	is.True(ok)

	rr.contentReader.N = 1

	hdr, err := r.Next()
	is.True(err != nil)
	is.True(hdr == nil)
}

func TestReader_ReadSize_DiscardEOF(t *testing.T) {
	var (
		is = is.New(t)
		er = errReader{
			errors: []error{
				io.EOF,
				io.EOF,
			},
		}
		r      = NewReader(&er)
		rr, ok = r.(*reader)
	)
	is.True(ok)

	rr.contentReader.N = 1

	hdr, err := r.Next()
	is.True(err != nil)
	is.True(hdr == nil)
}

func TestReader_ReadSizeEOF(t *testing.T) {
	var (
		is = is.New(t)
		er = errReader{
			errors: []error{
				nil,
				io.EOF,
			},
		}
		r = NewReader(&er)
	)

	hdr, err := r.Next()
	is.True(err != nil)
	is.True(hdr == nil)
}

func TestReader_ReadSizeErr(t *testing.T) {
	var (
		is = is.New(t)
		er = errReader{
			errors: []error{
				nil,
				errors.New("ERROR"),
			},
		}
		r = NewReader(&er)
	)

	hdr, err := r.Next()
	is.True(err != nil)
	is.True(hdr == nil)
}

type errReader struct {
	errors []error
}

func (er *errReader) Read(b []byte) (int, error) {
	var err error

	err, er.errors = er.errors[0], er.errors[1:]
	if err == nil {
		return len(b), nil
	}

	return 0, err
}
