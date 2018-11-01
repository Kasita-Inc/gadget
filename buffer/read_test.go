package buffer

import (
	"encoding/binary"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testReadable struct {
	bytecount  int
	bytesSent  int
	errorAfter int
	prefix     int
	err        error
	bytes      []byte
}

func NRandomBytes(n int) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func newBufferedReadable(bytecount int, errorAfter int, err error) *testReadable {
	return &testReadable{
		bytecount:  bytecount,
		errorAfter: errorAfter,
		err:        err,
		bytesSent:  0,
		bytes:      NRandomBytes(bytecount),
	}
}

func (br *testReadable) Read(b []byte) (n int, err error) {
	start := br.bytesSent
	end := br.bytesSent + len(b)
	if nil != br.err && end >= br.errorAfter {
		end = br.errorAfter
		err = br.err
	}
	if end > len(br.bytes) {
		end = len(br.bytes)
	}
	copy(b, br.bytes[start:end])
	br.bytesSent += end - start
	return end - start, err
}

var bufferReadTests = []*testReadable{
	newBufferedReadable(20, 0, nil),
	newBufferedReadable(20, 5, io.EOF),
	newBufferedReadable(1055, 0, nil),
	newBufferedReadable(1024, 1024, io.EOF),
}

func TestBufferedRead(t *testing.T) {
	a := assert.New(t)
	for i, test := range bufferReadTests {
		actual, err := Read(test)
		expected := test.bytes
		if nil != test.err {
			a.EqualError(err, test.err.Error(), "Expected error (%s) for test %d",
				test.err.Error(), i)
			expected = test.bytes[:test.errorAfter]
		}
		a.Equal(actual, expected, "Test Case %d", i)
	}
}

func prefixedRandomBytes(n int, prefix uint16) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	rand.Read(b)
	// fill in the prefix
	binary.LittleEndian.PutUint16(b, prefix)
	return b
}

func newPrefixBufferedReadable(bytecount, errorAfter int, prefix uint16, err error) *testReadable {
	return &testReadable{
		bytecount:  bytecount,
		errorAfter: errorAfter,
		err:        err,
		bytesSent:  0,
		bytes:      prefixedRandomBytes(bytecount, prefix),
		prefix:     int(prefix),
	}
}

var lePrefixRead16Tests = []*testReadable{
	newPrefixBufferedReadable(20, 0, 20, nil),
	newPrefixBufferedReadable(20, 20, 30, io.EOF),
	newPrefixBufferedReadable(40, 0, 30, nil),
	// on boundary
	newPrefixBufferedReadable(1024, 0, 1024, io.EOF),
	// over boundary
	newPrefixBufferedReadable(1055, 0, 1055, nil),
}

func TestPrefixRead16(t *testing.T) {
	a := assert.New(t)
	for i, test := range lePrefixRead16Tests {
		actual, err := LittleEndian.PrefixRead16(test)
		expected := test.bytes
		if nil != test.err {
			a.EqualError(err, test.err.Error(), "Expected error (%s) for test %d",
				test.err.Error(), i)
			if test.errorAfter > test.prefix {
				expected = test.bytes[:test.prefix]
			} else {
				expected = test.bytes[:test.errorAfter]
			}

		}
		if test.prefix < len(test.bytes) {
			expected = test.bytes[:test.prefix]
		}
		a.Equal(actual, expected, "Test Case %d", i)
	}
}
