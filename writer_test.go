package pitch

import (
	"bytes"
	"errors"
	"testing"

	"github.com/matryer/is"
)

func TestWriter_ErrClosed(t *testing.T) {
	var (
		is = is.New(t)
		w  = NewWriter(nil)
	)

	err := w.Close()
	is.True(errors.Is(err, ErrClosed))

	n, err := w.WriteHeader("", 1, nil)
	is.Equal(n, 0)
	is.True(errors.Is(err, ErrClosed))

	n, err = w.Write([]byte{10})
	is.True(errors.Is(err, ErrClosed))
	is.Equal(n, 0)
}

func TestWriter_NegativeContentSize(t *testing.T) {
	var (
		is  = is.New(t)
		buf = bytes.NewBuffer(nil)
		w   = NewWriter(buf)
	)

	n, err := w.WriteHeader("name", -1, nil)
	is.True(errors.Is(err, ErrInvalidSize))
	is.Equal(n, 0)
}

func TestWriter_WriteHeader_PadError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				errors.New("ERROR"),
			},
		}
		w = NewWriter(&ew)
	)

	n, err := w.WriteHeader("name", 1, nil)
	is.True(err != nil)
	is.Equal(n, 0)
}

func TestWriter_WriteHeader_WriteNameLengthError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				errors.New("ERROR"),
			},
		}
		w = NewWriter(&ew)
	)

	n, err := w.WriteHeader("name", 1, nil)
	is.True(err != nil)
	is.Equal(n, 0)
}
