package pitch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func WalkDirFunc(w *Writer, dir string) fs.WalkDirFunc {
	dir = filepath.Clean(dir)
	dirParent := filepath.Dir(dir)
	sep := string(filepath.Separator)
	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}

		// follow symlinks
		headerName := strings.TrimPrefix(path, dirParent+sep)
		size := info.Size()
		if entry.Type()&fs.ModeSymlink != 0 {
			path, err = os.Readlink(path)
			if err != nil {
				return fmt.Errorf("error reading symlink: %w", err)
			}
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("error getting file info: %w", err)
			}
			size = stat.Size()
		}

		if _, err := w.WriteHeader(headerName, size, nil); err != nil {
			return fmt.Errorf("error writing header (%s, %d): %w", headerName, info.Size(), err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}

		if _, err := io.Copy(w, file); err != nil {
			file.Close()
			return fmt.Errorf("error copying file [%s]: %w", path, err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("error closing file: %w", err)
		}

		return nil
	}
}

func ArchiveDir(dst io.WriteCloser, dir string) error {
	var pw = NewWriter(dst)
	defer pw.Close()

	return filepath.WalkDir(dir, WalkDirFunc(pw, dir))
}
