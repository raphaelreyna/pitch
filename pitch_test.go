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
					"a.txt": []byte("AAA"),
				},
			},
			{
				name: "multiple_files",
				files: map[string][]byte{
					"a.txt":     []byte("AAA"),
					"foo/b.txt": []byte("BBB"),
				},
			},
			{
				name: "long_name",
				files: map[string][]byte{
					strings.Repeat("a", 1024) + ".txt": []byte("AAAAAA"),
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
				_, err := w.WriteHeader(name, int64(len(contents)), nil)
				is.NoErr(err)
				_, err = w.Write(contents)
				is.NoErr(err)
			}
			err := w.Close()
			is.NoErr(err)

			var (
				unmatchedFiles = len(files)
				data           = make([]byte, buf.Len())
			)
			copy(data, buf.Bytes())
			toc, err := BuildTableOfContents(buf)
			is.NoErr(err)

			br := bytes.NewReader(data)
			for name, byteRange := range toc {
				_, err = br.Seek(byteRange.Start, 0)
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
				_, err := w.WriteHeader(name, int64(len(contents)), nil)
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

func TestBuildTableOfContents_CatReader(t *testing.T) {
	var (
		is = is.New(t)

		tests = []struct {
			name     string
			archives []map[*Header][]byte
		}{
			{
				name: "basic",
				archives: []map[*Header][]byte{
					NewContents(
						&Header{
							Name: "a.txt",
						},
						&Header{
							Name: "foo/b.txt",
						},
					),
					NewContents(
						&Header{
							Name: "c.txt",
						},
						&Header{
							Name: "foo/d.txt",
						},
					),
					NewContents(
						&Header{
							Name: "e.txt",
						},
						&Header{
							Name: "foo/f.txt",
						},
					),
				},
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				is      = is.New(t)
				readers = make([]Reader, len(test.archives))
				err     error
			)
			for i, files := range test.archives {
				readers[i], err = NewReaderFromMap(files)
				if _, ok := readers[i].(internalReader); !ok {
					is.Equal(ok, true)
				}
				is.NoErr(err)
			}

			archiveBuf := make([]byte, 0)
			for _, reader := range readers {
				rr, ok := reader.(*inMemoryReader)
				is.Equal(ok, true)
				archiveBuf = append(archiveBuf, rr.buf.Bytes()...)
			}

			r := Cat(readers...)

			toc, err := BuildTableOfContents(r)
			is.NoErr(err)
			is.True(toc != nil)

			for _, archive := range test.archives {
				for hdr, fileContents := range archive {
					br, ok := toc[hdr.Name]
					is.Equal(ok, true)
					extractedData := archiveBuf[br.Start:br.End]
					is.Equal(extractedData, fileContents)
				}
			}
		})
	}
}
