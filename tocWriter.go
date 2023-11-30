package pitch

import (
	"fmt"
	"io"
)

type ByteRange struct {
	Start int64
	End   int64
}

// HeaderItem is a struct that contains information about a file.
// In each archive file, a file's header is followed by its content.
type HeaderItem struct {
	Name string `json:"name" yaml:"name"`
	// Size is the size of the file content in bytes.
	Size uint64 `json:"size" yaml:"size"`
	// Data is a user-defined map of key-value pairs.
	Data map[string][]string `json:"data,omitempty" yaml:"data,omitempty"`
	// Start is the byte offset of the file content.
	Start int64 `json:"start" yaml:"start"`
	// End is the byte offset of the end of the file content.
	End int64 `json:"end" yaml:"end"`
}

// TableOfContents is a map of file names to HeaderItems.
type TableOfContents map[string]*HeaderItem

func TableToList[T ListOfContents | ListOfContentsByName | ListOfContentsByLocation](toc TableOfContents) T {
	var (
		loc = make(T, 0, len(toc))
	)

	for _, v := range toc {
		loc = append(loc, v)
	}

	return loc
}

type ListOfContents []*HeaderItem

type ListOfContentsByName []*HeaderItem

func (loc ListOfContentsByName) Len() int {
	return len(loc)
}

func (loc ListOfContentsByName) Less(i, j int) bool {
	return loc[i].Name < loc[j].Name
}

func (loc ListOfContentsByName) Swap(i, j int) {
	loc[i], loc[j] = loc[j], loc[i]
}

type ListOfContentsByLocation []*HeaderItem

func (loc ListOfContentsByLocation) Len() int {
	return len(loc)
}

func (loc ListOfContentsByLocation) Less(i, j int) bool {
	return loc[i].Start < loc[j].Start
}

func (loc ListOfContentsByLocation) Swap(i, j int) {
	loc[i], loc[j] = loc[j], loc[i]
}

type TOCWriter struct {
	contentLength int64
	w             io.Writer
	toc           TableOfContents
	offset        int64
}

func NewTOCWriter(w io.Writer) *TOCWriter {
	return &TOCWriter{
		w:   w,
		toc: make(TableOfContents),
	}
}

func (wtr *TOCWriter) WriteHeader(name string, contentLength int64, data map[string][]string) error {
	var w = wtr.w
	if w == nil {
		return ErrClosed
	}
	if contentLength < 1 {
		return ErrInvalidSize
	}

	if err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	hdr := Header{
		Name: name,
		Size: uint64(contentLength),
		Data: data,
	}
	payload := EncodeHeader(hdr)
	wtr.offset += int64(len(payload))
	_, err := w.Write(payload)
	if err != nil {
		return err
	}

	wtr.contentLength = contentLength
	wtr.toc[name] = &HeaderItem{
		Name:  name,
		Size:  uint64(contentLength),
		Data:  nil,
		Start: wtr.offset,
		End:   wtr.offset + contentLength,
	}

	return nil
}

func (wtr *TOCWriter) Write(b []byte) (int, error) {
	var w = wtr.w
	if w == nil {
		return 0, ErrClosed
	}

	var (
		n  = int64(len(b))
		cl = wtr.contentLength
	)

	if cl == 0 {
		return 0, ErrWriteTooLong
	}

	var isTooLong bool
	if cl = wtr.contentLength; cl < n {
		n = cl
		isTooLong = true
	}

	m, err := w.Write(b[:n])
	if err != nil {
		return m, err
	}

	var m64 = int64(m)
	wtr.offset += m64
	wtr.contentLength -= m64

	if isTooLong {
		err = ErrWriteTooLong
	}

	return m, err
}

func (wtr *TOCWriter) Close() error {
	if wtr.w == nil {
		return ErrClosed
	}

	if err := wtr.pad(); err != nil {
		return fmt.Errorf("error padding file: %w", err)
	}

	wtr.w = nil

	return nil
}

func (wtr *TOCWriter) pad() error {
	var (
		cl     = wtr.contentLength
		_, err = wtr.w.Write(make([]byte, cl))
	)

	wtr.offset += int64(cl)

	return err
}

func (wtr *TOCWriter) TableOfContents() TableOfContents {
	return wtr.toc
}
