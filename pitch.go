package pitch

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func BuildTableOfContents(v interface{}) (TableOfContents, error) {
	switch x := v.(type) {
	case []byte:
		var buf = bytes.NewReader(x)
		return buildTableOfContentsFromSeeker(buf)
	case io.ReadSeeker:
		return buildTableOfContentsFromSeeker(x)
	case io.Reader:
		return buildTableOfContentsFromReader(x)
	}

	return nil, errors.New("expected []byte, io.Reader or io.ReadSeeker")
}

func buildTableOfContentsFromReader(r io.Reader) (TableOfContents, error) {
	var (
		toc    = make(TableOfContents)
		pr     = NewReader(r)
		offset int64
	)

	for {
		var bytesRead int64

		nameSize, br, err := pr.readSize()
		if errors.Is(err, io.EOF) {
			return toc, nil
		}
		if err != nil {
			return nil, err
		}

		bytesRead += br

		var nameBuf = make([]byte, nameSize)
		_, err = io.ReadFull(pr.r, nameBuf)
		if err != nil {
			return nil, err
		}

		bytesRead += nameSize

		contentSize, br, err := pr.readSize()
		if err != nil {
			return nil, err
		}

		bytesRead += br
		offset += bytesRead

		_, err = io.CopyN(io.Discard, r, contentSize)
		if err != nil {
			return nil, err
		}

		toc[string(nameBuf)] = ByteRange{
			Start: offset,
			End:   offset + contentSize,
		}

		offset += contentSize
	}
}

func buildTableOfContentsFromSeeker(r io.ReadSeeker) (TableOfContents, error) {
	var (
		toc    = make(TableOfContents)
		pr     = NewReader(r)
		offset int64
	)

	for {
		var bytesRead int64

		nameSize, br, err := pr.readSize()
		if errors.Is(err, io.EOF) {
			return toc, nil
		}
		if err != nil {
			return nil, err
		}

		bytesRead += br

		var nameBuf = make([]byte, nameSize)
		_, err = io.ReadFull(pr.r, nameBuf)
		if err != nil {
			return nil, err
		}

		bytesRead += nameSize

		contentSize, br, err := pr.readSize()
		if err != nil {
			return nil, err
		}

		bytesRead += br
		offset += bytesRead

		if _, err := r.Seek(offset+contentSize, 0); err != nil {
			return nil, err
		}

		toc[string(nameBuf)] = ByteRange{
			Start: offset,
			End:   offset + contentSize,
		}

		offset += contentSize
	}
}

func ArchiveDir(dst io.WriteCloser, dir string) error {
	var pw = NewWriter(dst)
	defer pw.Close()

	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if err := pw.WriteHeader(path, info.Size()); err != nil {
			return err
		}

		file, err := os.Open(filepath.Join(dir, path))
		if err != nil {
			return err
		}

		if _, err := io.Copy(pw, file); err != nil {
			file.Close()

			return err
		}

		if err := file.Close(); err != nil {
			return err
		}

		return nil
	})
}
