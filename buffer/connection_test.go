package buffer

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/generator"
	gnet "github.com/Kasita-Inc/gadget/net"
)

type MockRWC struct {
	gnet.MockConn
	readThis []byte
	readErr  error
	writeErr error
}

func (mock *MockRWC) Close() error {
	return nil
}

func (mock *MockRWC) Read(data []byte) (int, error) {
	i := 0
	if nil != mock.readThis {
		copy(data, mock.readThis)
		i = len(mock.readThis)
	}
	return i, mock.readErr
}

func (mock *MockRWC) Write(data []byte) (int, error) {
	return len(data), mock.writeErr
}

type MockMarshaller struct {
	unmarshalError error
	marshalError   error
}

func (mm *MockMarshaller) MarshalBinary() ([]byte, error) {
	return generator.Bytes(10), mm.marshalError
}

func (mm *MockMarshaller) UnmarshalBinary(data []byte) error {
	return mm.unmarshalError
}

type mockErrorHandler struct{}

func (h *mockErrorHandler) HandleError(body []byte, err error) error {
	return errors.New("error handled")
}

const TO = time.Second

func TestNewConnection(t *testing.T) {
	assert := assert.New(t)
	connection := NewConnection(&MockRWC{}, true, TO, TO)
	assert.NotNil(connection)
}

func TestReadOnClosed(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{readThis: []byte{0}}
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	mu := &MockMarshaller{}
	assert.NoError(connection.Read(mu))
	connection.Close()
	assert.EqualError(connection.Read(mu), NewNetworkError(NewReadOnClosedError(), connection.GetConnectionInfo()).Error())
}

func TestReadEOFCloses(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{readThis: []byte{0, 2, 3, 4}}
	connection := NewConnection(rwc, true, TO, TO)
	assert.NotNil(connection)
	rwc.readErr = io.EOF
	mu := &MockMarshaller{}
	assert.NoError(connection.Read(mu))
	assert.EqualError(connection.Read(mu), NewNetworkError(NewReadOnClosedError(), connection.GetConnectionInfo()).Error())
}

func TestReadError(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{}
	expected := errors.New(generator.String(20))
	rwc.readErr = expected
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	mu := &MockMarshaller{}
	assert.EqualError(connection.Read(mu), NewNetworkError(expected, connection.GetConnectionInfo()).Error())
	assert.EqualError(connection.Read(mu), NewNetworkError(NewReadOnClosedError(), connection.GetConnectionInfo()).Error())
}

func TestReadUnmarshalError(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{readThis: []byte{0, 1, 2}}
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	expected := errors.New(generator.String(20))
	mu := &MockMarshaller{
		unmarshalError: expected,
	}
	assert.EqualError(connection.Read(mu), expected.Error())
}

func TestWriteOnClosed(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{}
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	connection.Close()
	mu := &MockMarshaller{}
	assert.EqualError(connection.Write(mu), NewNetworkError(NewWriteOnClosedError(), connection.GetConnectionInfo()).Error())
}

func TestWriteFailOnMarshal(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{}
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	expected := errors.New(generator.String(20))
	mu := &MockMarshaller{
		marshalError: expected,
	}
	assert.EqualError(connection.Write(mu), expected.Error())
}

func TestWriteFailOnWrite(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{}
	expected := errors.New(generator.String(20))
	rwc.writeErr = expected
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	mu := &MockMarshaller{}
	assert.EqualError(connection.Write(mu), NewNetworkError(expected, connection.GetConnectionInfo()).Error())
	assert.EqualError(connection.Read(mu), NewNetworkError(NewReadOnClosedError(), connection.GetConnectionInfo()).Error())
}

func TestWrite(t *testing.T) {
	assert := assert.New(t)
	rwc := &MockRWC{}
	connection := NewConnection(rwc, false, TO, TO)
	assert.NotNil(connection)
	mu := &MockMarshaller{}
	assert.NoError(connection.Write(mu))
}

func TestGetConnectionInfo(t *testing.T) {
	assert := assert.New(t)
	conn := &gnet.MockConn{RAddress: &gnet.MockAddr{SNetwork: "tcp", Address: "192.168.1.1"},
		LAddress: &gnet.MockAddr{SNetwork: "tcp", Address: "127.0.0.1"}}
	connection := NewConnection(conn, false, TO, TO)
	assert.NotNil(connection)
	assert.Equal(connection.GetConnectionInfo(), "MUBTO Connection Info(Local: 'tcp:127.0.0.1' RemoteAddress: 'tcp:192.168.1.1')")
}

func TestIsCrawler(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		bytes    []byte
		expected bool
	}{
		{
			bytes:    []byte{71, 69, 84, 32, 47, 114, 111, 98, 111, 116, 115, 46, 116, 120, 116, 32, 72, 84, 84, 80, 47, 49, 46, 49, 13, 10, 72, 111, 115, 116, 58, 32, 112, 114, 111, 100, 45, 107, 97, 115, 45, 110, 108, 98, 45, 102, 108, 117, 120, 45, 98, 52, 98, 102, 102, 50, 98, 52, 52, 98, 97, 48, 53, 49, 52, 97, 46, 101, 108, 98, 46, 117, 115, 45, 101, 97, 115, 116, 45, 49, 46, 97, 109, 97, 122, 111, 110, 97, 119, 115, 46, 99, 111, 109, 58, 49, 48, 53, 52, 57, 13, 10, 67, 111, 110, 110, 101, 99, 116, 105, 111, 110, 58, 32, 75, 101, 101, 112, 45, 65, 108, 105, 118, 101, 13, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 58, 32, 77, 111, 122, 105, 108, 108, 97, 47, 53, 46, 48, 32, 40, 99, 111, 109, 112, 97, 116, 105, 98, 108, 101, 59, 32, 89, 97, 110, 100, 101, 120, 66, 111, 116, 47, 51, 46, 48, 59, 32, 43, 104, 116, 116, 112, 58, 47, 47, 121, 97, 110, 100, 101, 120, 46, 99, 111, 109, 47, 98, 111, 116, 115, 41, 13, 10, 70, 114, 111, 109, 58},
			expected: true,
		},
		{
			bytes:    []byte{72, 69, 84, 32, 47, 114, 111, 98, 111, 116, 115, 46, 116, 120, 116, 32, 72, 84, 84, 80, 47, 49, 46, 49, 13, 10, 72, 111, 115, 116, 58, 32, 112, 114, 111, 100, 45, 107, 97, 115, 45, 110, 108, 98, 45, 102, 108, 117, 120, 45, 98, 52, 98, 102, 102, 50, 98, 52, 52, 98, 97, 48, 53, 49, 52, 97, 46, 101, 108, 98, 46, 117, 115, 45, 101, 97, 115, 116, 45, 49, 46, 97, 109, 97, 122, 111, 110, 97, 119, 115, 46, 99, 111, 109, 58, 49, 48, 53, 52, 57, 13, 10, 67, 111, 110, 110, 101, 99, 116, 105, 111, 110, 58, 32, 75, 101, 101, 112, 45, 65, 108, 105, 118, 101, 13, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 58, 32, 77, 111, 122, 105, 108, 108, 97, 47, 53, 46, 48, 32, 40, 99, 111, 109, 112, 97, 116, 105, 98, 108, 101, 59, 32, 89, 97, 110, 100, 101, 120, 66, 111, 116, 47, 51, 46, 48, 59, 32, 43, 104, 116, 116, 112, 58, 47, 47, 121, 97, 110, 100, 101, 120, 46, 99, 111, 109, 47, 98, 111, 116, 115, 41, 13, 10, 70, 114, 111, 109, 58},
			expected: false,
		},
	}
	for _, test := range cases {
		assert.Equal(IsCrawler(test.bytes), test.expected)
	}
}
