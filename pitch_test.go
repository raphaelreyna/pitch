package pitch

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestBuildTableOfContents(t *testing.T) {
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

				w = NewWriter(buf)
			)
			for name, contents := range files {
				err := w.WriteHeader(name, int64(len(contents)))
				is.NoErr(err)
				_, err = w.Write(contents)
				is.NoErr(err)
			}
			err := w.Close()
			is.NoErr(err)

			toc, err := BuildTableOfContents(buf.Bytes())
			is.NoErr(err)

			var (
				unmatchedFiles = len(files)
				br             = bytes.NewReader(buf.Bytes())
			)

			for name, byteRange := range toc {
				_, err := br.Seek(byteRange.Start, 0)
				is.NoErr(err)

				var cbuf = make([]byte, byteRange.End-byteRange.Start)
				_, err = io.ReadFull(br, cbuf)
				is.NoErr(err)

				content, ok := files[name]
				is.True(ok)

				is.Equal(content, cbuf)

				unmatchedFiles--
			}

			is.Equal(unmatchedFiles, 0)
		})
	}
}

func TestBuildTableOfContents_fromReader(t *testing.T) {
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

				w = NewWriter(buf)
			)
			for name, contents := range files {
				err := w.WriteHeader(name, int64(len(contents)))
				is.NoErr(err)
				_, err = w.Write(contents)
				is.NoErr(err)
			}
			err := w.Close()
			is.NoErr(err)

			toc, err := BuildTableOfContents(bytes.NewBuffer(buf.Bytes()))
			is.NoErr(err)

			var (
				unmatchedFiles = len(files)
				br             = bytes.NewReader(buf.Bytes())
			)

			for name, byteRange := range toc {
				_, err := br.Seek(byteRange.Start, 0)
				is.NoErr(err)

				var cbuf = make([]byte, byteRange.End-byteRange.Start)
				_, err = io.ReadFull(br, cbuf)
				is.NoErr(err)

				content, ok := files[name]
				is.True(ok)

				is.Equal(content, cbuf)

				unmatchedFiles--
			}

			is.Equal(unmatchedFiles, 0)
		})
	}
}
