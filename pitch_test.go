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
				r              = NewReader(br)
			)

			for name, loc := range toc {
				_, err := br.Seek(loc, 0)
				is.NoErr(err)

				readName, readSize, err := r.Next()
				is.NoErr(err)

				is.Equal(readName, name)

				content, ok := files[readName]
				is.True(ok)

				is.Equal(readSize, int64(len(content)))

				readContent, err := io.ReadAll(r)
				is.NoErr(err)

				is.Equal(content, readContent)

				unmatchedFiles--
			}

			is.Equal(unmatchedFiles, 0)
		})
	}
}
