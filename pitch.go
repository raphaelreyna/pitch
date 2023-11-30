package pitch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func BuildTableOfContents(v any) (TableOfContents, error) {
	switch x := v.(type) {
	case []byte:
		var buf = bytes.NewReader(x)
		return buildTableOfContentsFromReader(NewReader(buf))
	case Reader:
		return buildTableOfContentsFromReader(x)
	case io.Reader:
		return buildTableOfContentsFromReader(NewReader(x))
	}

	return nil, errors.New("expected []byte, Reader or io.Reader")
}

func buildTableOfContentsFromReader(r Reader) (TableOfContents, error) {
	var (
		toc    = make(TableOfContents)
		offset int64
	)

	for {
		hdr, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if len(toc) == 0 {
					return nil, io.EOF
				}
				return toc, nil
			}
			return nil, fmt.Errorf("error reading header: %w", err)
		}

		headerSize := int64(EncodedHeaderSize(hdr.Name, hdr.Size, hdr.Data))
		filesize := headerSize + int64(hdr.Size)

		toc[hdr.Name] = &HeaderItem{
			Name:  hdr.Name,
			Size:  hdr.Size,
			Data:  hdr.Data,
			Start: offset + headerSize,
			End:   offset + filesize,
		}
		offset += filesize
	}
}

func WalkFunc(w *Writer, dir string) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if _, err := w.WriteHeader(path, info.Size(), nil); err != nil {
			return err
		}

		file, err := os.Open(filepath.Join(dir, path))
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, file); err != nil {
			file.Close()

			return err
		}

		if err := file.Close(); err != nil {
			return err
		}

		return nil
	}
}

func ArchiveDir(dst io.WriteCloser, dir string) error {
	var pw = NewWriter(dst)
	defer pw.Close()

	return filepath.Walk(dir, WalkFunc(pw, dir))
}
