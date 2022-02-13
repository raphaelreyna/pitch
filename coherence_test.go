package pitch

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestCoherence(t *testing.T) {
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

			var r = NewReader(buf)

			var unmatchedFiles = len(files)
			for 0 < unmatchedFiles {
				name, size, err := r.Next()
				is.NoErr(err)

				fmt.Printf("name: %s\n", name)
				expectedContents, nameExists := files[name]
				is.True(nameExists)

				contents, err := io.ReadAll(r)
				is.NoErr(err)

				is.Equal(int(size), len(expectedContents))
				is.Equal(contents, expectedContents)

				unmatchedFiles--
			}
		})
	}
}
