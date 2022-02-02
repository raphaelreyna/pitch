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
		r = NewReader(&er)
	)

	r.contentReader.N = 1

	name, err := r.Next()
	is.True(err != nil)
	is.Equal(name, "")
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
		r = NewReader(&er)
	)

	r.contentReader.N = 1

	name, err := r.Next()
	is.True(err != nil)
	is.Equal(name, "")
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

	name, err := r.Next()
	is.True(err != nil)
	is.Equal(name, "")
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

	name, err := r.Next()
	is.True(err != nil)
	is.Equal(name, "")
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
