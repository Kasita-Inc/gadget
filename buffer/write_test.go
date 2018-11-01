package buffer

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testWritable struct {
	bytes      []byte
	bufferSize int
	errorAfter int
	err        error
}

func newTestWritable(bufferSize, errorAfter int, err error) *testWritable {
	return &testWritable{
		bytes:      []byte{},
		bufferSize: bufferSize,
		errorAfter: errorAfter,
		err:        err,
	}
}

func (tw *testWritable) Write(b []byte) (n int, err error) {
	n = tw.bufferSize
	if nil != tw.err {
		err = tw.err
		n = tw.errorAfter - len(tw.bytes)
	}
	if n > len(b) {
		n = len(b)
	}
	tw.bytes = append(tw.bytes, b[:n]...)
	return n, err
}

var testCases = []*testWritable{
	newTestWritable(20, 0, nil),
	newTestWritable(1024, 0, nil),
	newTestWritable(20, 10, io.EOF),
}

func TestWritable(t *testing.T) {
	assert := assert.New(t)
	for i, test := range testCases {
		count := 210
		expected := NRandomBytes(count)
		err := Write(test, expected)
		if nil != test.err {
			assert.EqualError(err, test.err.Error(), "expected error for test case %d", i)
			expected = expected[:test.errorAfter]
		}
		assert.Equal(expected, test.bytes, "test case %d", i)
	}
}
