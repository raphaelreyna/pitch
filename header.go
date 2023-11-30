package pitch

import (
	"bytes"
	"fmt"
	"io"
)

type SizeType uint8

const (
	NameSize = SizeType(iota)
	ContentSize
	DataNameSize
	DataValueSize
)

type Size struct {
	Type  SizeType
	Value uint64
}

func resizeBuffer(buf *bytes.Buffer, newSize uint64) {
	delta := int(newSize) - buf.Available()
	if 0 < delta {
		buf.Grow(int(delta))
	}
}

// DecodeSize reads the next size from r.
// The argument buf is used as a temporary buffer that must be at least 1 byte long.
func DecodeSize(r io.Reader, buf []byte) (s Size, err error) {
	bbuf := buf[:1]
	if len(bbuf) < 1 {
		bbuf = make([]byte, 1)
	}
	_, err = r.Read(bbuf)
	if err != nil {
		return
	}

	headerByte := bbuf[0]
	s.Type = SizeType(headerByte >> 6)
	encodedValueData := uint8(headerByte & 0b00111110)
	isFinalByte := headerByte&0b00000001 == 0b00000001
	s.Value = uint64(encodedValueData) >> 1

	if isFinalByte {
		return
	}

	shift := 5
	for i := 1; !isFinalByte; i++ {
		_, err = r.Read(bbuf)
		if err != nil {
			return
		}
		b := bbuf[0]
		isFinalByte = b&0b00000001 == 0b00000001
		b = b >> 1

		s.Value |= uint64(b) << shift
		shift += 7
	}

	return
}

// Header represents a file header in a pitch archive file.
// The header is followed by the file content.
type Header struct {
	// Name is the name of the file.
	Name string
	// Size is the size of the file content.
	Size uint64
	// Data is a user-defined map of key-value pairs.
	// This data is opaque to the pitch package.
	Data map[string][]string
}

// DecodeHeader reads the next header from r.
func DecodeHeader(r io.Reader) (*Header, error) {
	buf := bytes.NewBuffer(make([]byte, 1<<8))
	h := Header{
		Data: make(map[string][]string),
	}
	defer func() {
		if len(h.Data) == 0 {
			h.Data = nil
		}
	}()

	s, err := DecodeSize(r, buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error reading name size: %w", err)
	}

	if s.Type != NameSize {
		return nil, fmt.Errorf("expected name size, got %d", s.Type)
	}
	if s.Value == 0 {
		return nil, io.EOF
	}

	bbuf := make([]byte, s.Value)

	buf.Reset()
	resizeBuffer(buf, s.Value)
	_, err = r.Read(bbuf)
	if err != nil {
		return nil, fmt.Errorf("error reading name: %w", err)
	}
	h.Name = string(bbuf)
	if h.Name == "" {
		return nil, fmt.Errorf("unexpected empty name")
	}

	name := ""
	value := ""

	for done := false; !done; {
		buf.Reset()
		s, err = DecodeSize(r, buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("error reading size: %w", err)
		}

		switch s.Type {
		case NameSize:
			return nil, fmt.Errorf("unexpected name size")
		case ContentSize:
			done = true
			h.Size = s.Value
		case DataNameSize:
			if name != "" {
				h.Data[name] = append(h.Data[name], value)
			}
			name = ""

			buf.Reset()
			resizeBuffer(buf, s.Value)
			data := buf.Bytes()[:s.Value]
			if _, err := io.ReadFull(r, data); err != nil {
				return nil, fmt.Errorf("error reading header byte: %w", err)
			}
			name = string(data)
		case DataValueSize:
			if name == "" {
				return nil, fmt.Errorf("unexpected value size")
			}

			buf.Reset()
			resizeBuffer(buf, s.Value)
			data := buf.Bytes()[:s.Value]
			if _, err := io.ReadFull(r, data); err != nil {
				return nil, fmt.Errorf("error reading header byte: %w", err)
			}
			h.Data[name] = append(h.Data[name], string(data))
		}
	}

	return &h, nil
}

// EncodeSize encodes the given size and its type into a byte slice.
func EncodeSize(t SizeType, x uint64) []byte {
	bc := ByteCount(x)
	bytes := make([]byte, bc)

	// grab the lowest 5 bits
	b := uint8(x & 0b00011111)
	x >>= 5
	b |= uint8(t) << 5
	bytes[0] = b << 1
	if bc == 1 {
		bytes[0] |= 0b00000001
		return bytes
	}

	for i := bc - 1; 0 < i; i-- {
		b = uint8(x & 0b01111111)
		x >>= 7
		bytes[bc-i] = b << 1
	}

	bytes[len(bytes)-1] |= 0b00000001
	return bytes
}

// ByteCount returns the number of bytes required to encode x.
func ByteCount(x uint64) int {
	if x == 0 {
		return 1 // Handle 0 as a special case to return 1 byte
	}

	bitCount := 0
	for x > 0 {
		bitCount++
		x >>= 1
	}

	// We can stuff 5 bits into the header byte.
	// If we need more than that, we need to add more bytes (each of which can stuff 7 bits).
	if bitCount <= 5 {
		return 1
	}

	return 1 + ((bitCount - 6) / 7) + 1
}

// EncodedHeaderSize returns the number of bytes required to encode a header with the given name, size and data.
func EncodedHeaderSize(name string, size uint64, data map[string][]string) uint64 {
	var (
		nameSize     = uint64(len(name))
		nameSizeSize = uint64(ByteCount(nameSize))
		sizeSize     = uint64(ByteCount(size))
		dataSize     = uint64(0)
	)

	for k, v := range data {
		optionalNameSize := uint64(len(k))
		optionalNameSizeSize := uint64(ByteCount(optionalNameSize))
		ons := optionalNameSizeSize + optionalNameSize
		for _, s := range v {
			// if a key has multiple values, then it was encoded multiple times
			dataSize += ons
			dataSize += uint64(len(s))
		}
		// if a key has no values, then it was encoded once
		if len(v) == 0 {
			dataSize += ons
		}
	}

	return nameSizeSize + nameSize + dataSize + sizeSize
}

// EncodeHeader encodes the given header into a byte slice.
func EncodeHeader(h Header) []byte {
	var (
		buf      = bytes.NewBuffer(nil)
		nameSize = uint64(len(h.Name))
	)

	buf.Write(EncodeSize(NameSize, nameSize))
	buf.WriteString(h.Name)

	for k, v := range h.Data {
		optionalNameSize := uint64(len(k))
		buf.Write(EncodeSize(DataNameSize, optionalNameSize))
		buf.WriteString(k)
		for _, s := range v {
			buf.Write(EncodeSize(DataValueSize, uint64(len(s))))
			buf.WriteString(s)
		}
	}

	buf.Write(EncodeSize(ContentSize, h.Size))

	return buf.Bytes()
}
