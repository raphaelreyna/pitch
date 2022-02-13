package main

import (
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/raphaelreyna/pitch"
)

var (
	create   = flag.Bool("c", false, "create")
	extract  = flag.Bool("x", false, "extract")
	fileName = flag.String("f", "", "file name")
	dir      = flag.String("C", "", "directory")
)

func main() {
	flag.Parse()

	if *create && *extract {
		log.Fatalln(`pitch: You may not specify more than one '-xc' option`)
	}

	switch {
	case *create:
		if err := createFn(); err != nil {
			log.Fatalln(err)
		}
	case *extract:
		if err := extractFn(); err != nil {
			log.Fatalln(err)
		}
	}
}

func createFn() error {
	var (
		inDir   = flag.Args()[0]
		filesys = os.DirFS(inDir)
		w       = os.Stdout
	)

	if fileName != nil {
		var err error
		*fileName, err = filepath.Abs(*fileName)

		if err != nil {
			return err
		}

		file, err := os.Create(*fileName)
		if err != nil {
			return err
		}

		defer file.Close()

		w = file
	}

	var pw = pitch.NewWriter(w)
	var err = fs.WalkDir(filesys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		var fullpath = filepath.Join(inDir, path)
		fullpath, err = filepath.Abs(fullpath)
		if err != nil {
			return err
		}
		if fileName != nil {
			if fullpath == *fileName {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if err := pw.WriteHeader(path, info.Size()); err != nil {
			return err
		}

		file, err := os.Open(fullpath)
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = io.Copy(pw, file)

		return err
	})

	if err != nil {
		return err
	}

	return pw.Close()
}

func extractFn() error {
	var (
		outDir = "."
		r      = os.Stdin
	)

	if dir != nil {
		outDir = *dir
	}

	if fileName != nil {
		file, err := os.Open(*fileName)
		if err != nil {
			return err
		}

		defer file.Close()

		r = file
	}

	var pr = pitch.NewReader(r)

	for {
		name, _, err := pr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		var (
			fullpath  = filepath.Join(outDir, name)
			parentDir = filepath.Dir(fullpath)
		)

		if err := os.MkdirAll(parentDir, 0777); err != nil {
			return err
		}

		err = func() error {
			file, err := os.Create(fullpath)
			if err != nil {
				return err
			}

			defer file.Close()

			_, err = io.Copy(file, pr)

			return err
		}()

		if err != nil {
			return err
		}
	}
}
