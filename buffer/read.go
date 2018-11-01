package buffer

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/Kasita-Inc/gadget/errors"
)

const (
	read  = "READ"
	write = "WRITE"
)

// Read a readable instance by filling a buffer of constant size until less bytes are read than requested.
// Returns all byte concatenated together with non-read bytes truncated.
// NOTE: This function DOES return io.EOF errors which will always terminate the read AFTER the EOF, i.e. read bytes
// will still be appended when the EOF is returned.
func Read(readable io.Reader) ([]byte, error) {
	var buffer []byte
	var n int
	var err error
	received := make([]byte, 0)

	for {
		buffer = make([]byte, 512)
		n, err = readable.Read(buffer[:])
		if n > 0 {
			received = append(received, buffer[:n]...)
		}
		if n == 0 || n < len(buffer) {
			break
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				err = NewTimeoutError(read, len(received))
			}
			break
		}
	}

	return received, err
}

type littleEndian struct{}

// LittleEndian encapsulates read methods that assume a LittleEndian byte order for the length prefix.
var LittleEndian littleEndian

// PrefixRead16 reads from the passed reader assuming the first two bytes are a uint16 containing the
// length of the total payload. Extra bytes are truncated.
func (littleEndian) PrefixRead16(readable io.Reader) ([]byte, error) {
	if nil == readable {
		return nil, errors.New("buffer: readable was <nil>")
	}
	var buffer []byte
	var n int
	var length int
	var err error
	received := make([]byte, 0)
	nTotal := 0

	for {
		buffer = make([]byte, 512)
		n, err = readable.Read(buffer[:])
		nTotal += n
		if n > 0 {
			received = append(received, buffer[:n]...)
			if 0 == length && nTotal > 1 {
				length = int(binary.LittleEndian.Uint16(received))
			}
		}
		if 0 == n || (length > 0 && nTotal >= length) {
			break
		}
		if err != nil {

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				err = NewTimeoutError(read, len(received))
			}
			break
		}
	}
	if length < nTotal {
		received = received[:length]
	}
	return received, err
}
