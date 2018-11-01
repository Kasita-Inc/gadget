package buffer

import (
	"io"
	"net"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/intutil"
)

// BufferSizeBytes is the number of
const BufferSizeBytes = 512

// Write of the complete contents of data over the passed Writable.
func Write(writable io.Writer, data []byte) (err error) {
	var n int
	lenData := len(data)
	for bytesSent := 0; bytesSent < lenData; {
		end := intutil.Min(bytesSent+BufferSizeBytes, lenData)
		var bytes []byte
		if lenData == end {
			bytes = data[bytesSent:]
		} else {
			bytes = data[bytesSent:end]
		}
		n, err = writable.Write(bytes)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				err = NewTimeoutError(write, bytesSent)
			}
			break
		} else if n == 0 {
			return errors.New("write failed, no bytes written")
		}
		bytesSent += n
	}
	return
}
