package buffer

import (
	"net"

	"github.com/Kasita-Inc/gadget/collection"
	"github.com/Kasita-Inc/gadget/errors"
)

// MockConnection is used for testing
type MockConnection struct {
	ReadMessages  collection.Stack
	ReadFunc      func(mu MarshalUnmarshal) error
	WriteMessages collection.Stack
	WriteError    error
	IsClosed      bool
}

// NewMockConnection returns a MockConnection
func NewMockConnection() *MockConnection {
	return &MockConnection{
		ReadMessages:  collection.NewStack(),
		WriteMessages: collection.NewStack(),
	}
}

// GetConnection is a no-op
func (conn *MockConnection) GetConnection() net.Conn {
	return nil
}

// GetConnectionInfo is always blank for the mock connection
func (conn *MockConnection) GetConnectionInfo() string {
	return ""
}

// Closed sets the MockConnection IsClosed to true
func (conn *MockConnection) Closed() bool {
	return conn.IsClosed
}

// Read either uses the ReadFunc or pops a message from the ReadMessages stack
func (conn *MockConnection) Read(mu MarshalUnmarshal) errors.TracerError {
	if nil != conn.ReadFunc {
		return errors.Wrap(conn.ReadFunc(mu))
	}
	v, e := conn.ReadMessages.Pop()
	var message MarshalUnmarshal
	var err error
	var ok bool
	if nil == e {
		message, ok = v.(MarshalUnmarshal)
		if !ok {
			err, _ = v.(error)
		} else {
			data, _ := message.MarshalBinary()
			err = mu.UnmarshalBinary(data)
		}
	}
	return errors.Wrap(err)
}

// Write pushed the message into a Stack and returns the WriteError on the MockConnection
func (conn *MockConnection) Write(mu MarshalUnmarshal) errors.TracerError {
	conn.WriteMessages.Push(mu)
	return errors.Wrap(conn.WriteError)
}

// Close sets IsClosed to true on the MockConnection
func (conn *MockConnection) Close() errors.TracerError {
	conn.IsClosed = true
	return nil
}
