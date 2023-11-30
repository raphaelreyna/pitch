package pitch

import (
	"encoding/binary"
	"math"
	"unsafe"
)

var nativeEndian binary.ByteOrder

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

// Int64ByteCount returns the number of bytes required to encode a particular int64
// value in the pitch format header.
func EncodedInt64Size(x int64) int64 {
	digitCount := math.Log2(float64(x))
	digitCount = math.Floor(digitCount) + 1
	return int64(math.Ceil(digitCount / 7))
}

// EncodedStringSize returns the number of bytes required to encode a particular string
// value in the pitch format header.
// This includes both the initial size of the string and the string itself.
func EncodedStringSize(s string) int64 {
	header := EncodedInt64Size(int64(len(s)))
	return header + int64(len(s))
}

func HeaderSize(name string, size int64) int64 {
	return EncodedStringSize(name) + EncodedInt64Size(size)
}

// EncodedFileSize returns the number of bytes required to encode a particular file (including the header)
// into a pitch format archive.
func EncodedFileSize(name string, size int64) int64 {
	return HeaderSize(name, size) + size
}
