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

	n, err := w.WriteHeader("", 1)
	is.Equal(n, 0)
	is.True(errors.Is(err, ErrClosed))

	n, err = w.Write([]byte{10})
	is.True(errors.Is(err, ErrClosed))
	is.Equal(n, 0)
}

func TestWriter_ErrWriteTooLong(t *testing.T) {
	var (
		is  = is.New(t)
		buf = bytes.NewBuffer(nil)
		w   = NewWriter(buf)
	)

	n, err := w.WriteHeader("name", 1)
	is.NoErr(err)
	is.Equal(n, headerSize("name", 1))

	n, err = w.Write([]byte{10, 10})
	is.True(errors.Is(err, ErrWriteTooLong))
	is.Equal(n, 1)

	n, err = w.Write([]byte{10, 10})
	is.True(errors.Is(err, ErrWriteTooLong))
	is.Equal(n, 0)
}

func TestWriter_NegativeContentSize(t *testing.T) {
	var (
		is  = is.New(t)
		buf = bytes.NewBuffer(nil)
		w   = NewWriter(buf)
	)

	n, err := w.WriteHeader("name", -1)
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

	n, err := w.WriteHeader("name", 1)
	is.True(err != nil)
	is.Equal(n, 0)
}

func TestWriter_WriteHeader_WriteNameLengthError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				nil,
				errors.New("ERROR"),
			},
		}
		w = NewWriter(&ew)
	)

	n, err := w.WriteHeader("name", 1)
	is.True(err != nil)
	is.Equal(n, 0)
}

func TestWriter_WriteHeader_WriteNameError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				nil, nil,
				errors.New("ERROR"),
			},
		}
		w = NewWriter(&ew)
	)

	n, err := w.WriteHeader("name", 1)
	is.True(err != nil)
	is.Equal(n, sizeWriteSize(len("name")))
}

func TestWriter_Write_WriteError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				nil, nil, nil, nil,
				errors.New("ERROR"),
			},
		}
		w = NewWriter(&ew)
	)

	n, err := w.WriteHeader("name", 1)
	is.NoErr(err)
	is.Equal(n, headerSize("name", 1))

	n, err = w.Write([]byte{10})
	is.True(err != nil)
	is.Equal(0, n)
}

func TestWriter_Close_PadWriteError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				errors.New("ERROR"),
			},
		}
		w = NewWriter(&ew)
	)

	err := w.Close()
	is.True(err != nil)
}

func headerSize(name string, contentSize int) int {
	l := len(name)
	return sizeWriteSize(l) + l + sizeWriteSize(contentSize)
}
