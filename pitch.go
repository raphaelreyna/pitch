package pitch

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

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
