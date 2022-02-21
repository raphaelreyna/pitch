package pitch

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestTOCWriter_ErrClosed(t *testing.T) {
	var (
		is = is.New(t)
		w  = NewTOCWriter(nil)
	)

	err := w.Close()
	is.True(errors.Is(err, ErrClosed))

	err = w.WriteHeader("", 1)
	is.True(errors.Is(err, ErrClosed))

	n, err := w.Write([]byte{10})
	is.True(errors.Is(err, ErrClosed))
	is.Equal(n, 0)
}

func TestTOCWriter_ErrWriteTooLong(t *testing.T) {
	var (
		is  = is.New(t)
		buf = bytes.NewBuffer(nil)
		w   = NewTOCWriter(buf)
	)

	err := w.WriteHeader("name", 1)
	is.NoErr(err)

	n, err := w.Write([]byte{10, 10})
	is.True(errors.Is(err, ErrWriteTooLong))
	is.Equal(n, 1)

	n, err = w.Write([]byte{10, 10})
	is.True(errors.Is(err, ErrWriteTooLong))
	is.Equal(n, 0)
}

func TestTOCWriter_NegativeContentSize(t *testing.T) {
	var (
		is  = is.New(t)
		buf = bytes.NewBuffer(nil)
		w   = NewTOCWriter(buf)
	)

	err := w.WriteHeader("name", -1)
	is.True(errors.Is(err, ErrInvalidSize))
}

func TestTOCWriter_WriteHeader_PadError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				errors.New("ERROR"),
			},
		}
		w = NewTOCWriter(&ew)
	)

	err := w.WriteHeader("name", 1)
	is.True(err != nil)
}

func TestTOCWriter_WriteHeader_WriteNameLengthError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				nil,
				errors.New("ERROR"),
			},
		}
		w = NewTOCWriter(&ew)
	)

	err := w.WriteHeader("name", 1)
	is.True(err != nil)
}

func TestTOCWriter_WriteHeader_WriteNameError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				nil, nil,
				errors.New("ERROR"),
			},
		}
		w = NewTOCWriter(&ew)
	)

	err := w.WriteHeader("name", 1)
	is.True(err != nil)
}

func TestTOCWriter_Write_WriteError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				nil, nil, nil, nil,
				errors.New("ERROR"),
			},
		}
		w = NewTOCWriter(&ew)
	)

	err := w.WriteHeader("name", 1)
	is.NoErr(err)

	n, err := w.Write([]byte{10})
	is.True(err != nil)
	is.Equal(0, n)
}

func TestTOCWriter_Close_PadWriteError(t *testing.T) {
	var (
		is = is.New(t)
		ew = errWriter{
			errors: []error{
				errors.New("ERROR"),
			},
		}
		w = NewTOCWriter(&ew)
	)

	err := w.Close()
	is.True(err != nil)
}

type errWriter struct {
	errors []error
}

func (ew *errWriter) Write(b []byte) (int, error) {
	var err error

	err, ew.errors = ew.errors[0], ew.errors[1:]
	if err == nil {
		return len(b), nil
	}

	return 0, err
}

func TestTOCWriter_TableOfContents(t *testing.T) {
	var (
		is = is.New(t)

		tests = []struct {
			name  string
			files map[string][]byte
		}{
			{
				name: "basic",
				files: map[string][]byte{
					"a.txt": []byte("a.txt contents"),
				},
			},
			{
				name: "multiple_files",
				files: map[string][]byte{
					"a.txt":     []byte("a.txt contents"),
					"foo/b.txt": []byte("foo/b.txt contents"),
				},
			},
			{
				name: "long_name",
				files: map[string][]byte{
					strings.Repeat("a", 1024) + ".txt": []byte("a.txt contents"),
				},
			},
			{
				name: "long_contents",
				files: map[string][]byte{
					"a.txt": []byte(strings.Repeat("a", 4017)),
				},
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				is    = is.New(t)
				files = test.files

				buf = bytes.NewBuffer(nil)

				w = NewTOCWriter(buf)
			)
			for name, contents := range files {
				err := w.WriteHeader(name, int64(len(contents)))
				is.NoErr(err)
				_, err = w.Write(contents)
				is.NoErr(err)
			}
			err := w.Close()
			is.NoErr(err)

			var (
				toc  = w.TableOfContents()
				rbuf = bytes.NewReader(buf.Bytes())
				r    = NewReader(rbuf)
			)

			for name, content := range files {
				var loc, found = toc[name]
				is.True(found)
				rbuf.Seek(loc, 0)
				readName, readSize, err := r.Next()
				is.NoErr(err)
				is.Equal(readName, name)
				is.Equal(readSize, int64(len(content)))

				readContents, err := io.ReadAll(r)
				is.NoErr(err)
				is.Equal(readContents, content)
			}
		})
	}
}
