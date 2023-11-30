package pitch

import (
	"bytes"
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
			files map[*Header][]byte
		}{
			{
				name: "basic",
				files: NewContents(&Header{
					Name: "a.txt",
					Data: map[string][]string{
						"Content-Type": {"text/plain"},
					},
				}),
			},
			{
				name: "multiple_files",
				files: NewContents(&Header{
					Name: "a.txt",
				}, &Header{
					Name: "foo/b.txt",
				}),
			},
			{
				name: "long_name",
				files: NewContents(&Header{
					Name: strings.Repeat("a", 1024) + ".txt",
					Data: map[string][]string{
						"Content-Type": {"text/plain", "text/html"},
					},
				}),
			},
			{
				name: "long_contents",
				files: NewContents(&Header{
					Name: "a.txt",
					Size: 4017,
				}),
			},
		}
	)

	type fileBundle struct {
		hdr      *Header
		contents []byte
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				is          = is.New(t)
				files       = test.files
				filesByName = make(map[string]fileBundle, len(files))

				buf = bytes.NewBuffer(nil)

				w = NewWriter(buf)
			)

			for hdr, contents := range files {
				filesByName[hdr.Name] = fileBundle{
					hdr:      hdr,
					contents: contents,
				}
			}

			for hdr, contents := range files {
				_, err := w.WriteHeader(hdr.Name, int64(hdr.Size), hdr.Data)
				is.NoErr(err)
				_, err = w.Write(contents)
				is.NoErr(err)
			}
			err := w.Close()
			is.NoErr(err)

			var r = NewReader(buf)

			var unmatchedFiles = len(files)
			for 0 < unmatchedFiles {
				hdr, err := r.Next()
				is.NoErr(err)

				bundle, nameExists := filesByName[hdr.Name]
				is.True(nameExists)

				contents, err := io.ReadAll(r)
				is.NoErr(err)

				is.Equal(int(hdr.Size), len(bundle.contents))
				is.Equal(contents, bundle.contents)

				for k, v := range hdr.Data {
					bundleV, ok := bundle.hdr.Data[k]
					is.True(ok)
					for i, v := range v {
						is.Equal(v, bundleV[i])
					}
				}

				unmatchedFiles--
			}
		})
	}
}
