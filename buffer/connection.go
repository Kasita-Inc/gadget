package buffer

import (
	"encoding"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Kasita-Inc/gadget/errors"
)

// MarshalError is a marker for triggering logic based upon errors that occur during the marshal/unmarshal routines.
type MarshalError struct {
	trace     []string
	BaseError errors.TracerError
}

// NewMarshalErrorf from the passed format string and arguments
func NewMarshalErrorf(fmt string, arg ...interface{}) *MarshalError {
	return &MarshalError{BaseError: errors.New(fmt, arg...)}
}

// NewMarshalError with the passed base error
func NewMarshalError(err error) *MarshalError {
	merr, ok := err.(*MarshalError)
	if ok {
		return merr
	}
	if nil != err {
		return &MarshalError{BaseError: errors.Wrap(err)}
	}
	return nil
}

// Trace the stack for this errors base error.
func (err *MarshalError) Trace() []string {
	return err.BaseError.Trace()
}

func (err *MarshalError) Error() string {
	return err.BaseError.Error()
}

// NetworkError includes connection info.
type NetworkError struct {
	trace          []string
	BaseError      errors.TracerError
	ConnectionInfo string
}

// NewNetworkError with the passed base error and connection info.
func NewNetworkError(err error, connectionInfo string) errors.TracerError {
	if nil != err {
		return &NetworkError{BaseError: errors.Wrap(err), ConnectionInfo: connectionInfo}
	}
	return nil
}

// Trace the stack for this errors base error.
func (err *NetworkError) Trace() []string {
	return err.BaseError.Trace()
}

func (err *NetworkError) Error() string {
	return fmt.Sprintf("NetworkError (%s): %s", err.ConnectionInfo, err.BaseError.Error())
}

// ReadOnClosedError  is returned when an attempt to read is made on a nil connection.
type ReadOnClosedError struct{ trace []string }

func (err *ReadOnClosedError) Error() string {
	return "read called on closed connection"
}

// Trace returns the stack trace for the error
func (err *ReadOnClosedError) Trace() []string {
	return err.trace
}

// NewReadOnClosedError instantiates a ReadOnClosedError with a stack trace
func NewReadOnClosedError() errors.TracerError {
	return &ReadOnClosedError{trace: errors.GetStackTrace()}
}

// WriteOnClosedError  is returned when an attempt to write is made on a nil connection.
type WriteOnClosedError struct{ trace []string }

func (err *WriteOnClosedError) Error() string {
	return "write called on closed connection"
}

// Trace returns the stack trace for the error
func (err *WriteOnClosedError) Trace() []string {
	return err.trace
}

// NewWriteOnClosedError instantiates a WriteOnClosedError with a stack trace
func NewWriteOnClosedError() errors.TracerError {
	return &WriteOnClosedError{trace: errors.GetStackTrace()}
}

const (
	// ErrDisconnectNoDataMessage is a signal that the remote peer terminated the connection and sent no data
	ErrDisconnectNoDataMessage = "peer disconnected prior to sending data"
	// ErrHealthCheckMessage is a signal that the remote peer sent no data but did not terminate the connection
	ErrHealthCheckMessage = "health check errors should be ignored"
	// ErrCipherAuthenticationFailed is returned when the cipher on hand is not valid for the received message.
	ErrCipherAuthenticationFailed = "cipher: message authentication failed"
)

// NewErrDisconnectNoData is a signal that the remote peer terminated the connection and sent no data
func NewErrDisconnectNoData() errors.TracerError {
	return errors.New(ErrDisconnectNoDataMessage)
}

// NewErrHealthCheck is a signal that the remote peer sent no data
func NewErrHealthCheck() errors.TracerError {
	return errors.New(ErrHealthCheckMessage)
}

// Connection allows for two way exchange of messages.
type Connection interface {
	// Read and unmarshal into the passed MarshalUnmarshal from the connection.
	Read(mu MarshalUnmarshal) (err errors.TracerError)
	// Write the passed MarshalUnmarshal to the connection.
	Write(mu MarshalUnmarshal) errors.TracerError
	// Close this connection. Once closed this connection cannot be used further.
	Close() errors.TracerError
	// Closed returns a bool indicating that the underlying connection is closed.
	Closed() bool
	// GetConnection that is backing this abstraction.
	GetConnection() net.Conn
	// GetConnectionInfo in string format.
	GetConnectionInfo() string
}

// MarshalUnmarshal indicates a struct that can be marshalled to bytes and Unmarshalled from bytes.
type MarshalUnmarshal interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// NewConnection using the passed network connection.
func NewConnection(conn net.Conn, prefixRead bool, readTimeout time.Duration, writeTimeout time.Duration) Connection {
	mubto := &mubtoConnection{
		conn:         conn,
		prefixRead:   prefixRead,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
	mubto.setConnectionInfo()
	return mubto
}

// mubtoConnection marshalling unmarshalling buffered time out connection (MUBTO)
type mubtoConnection struct {
	conn           net.Conn
	connectionInfo string
	prefixRead     bool
	readTimeout    time.Duration
	writeTimeout   time.Duration
}

func (mubto *mubtoConnection) GetConnection() net.Conn {
	return mubto.conn
}

func (mubto *mubtoConnection) setConnectionInfo() {
	conn, ok := mubto.conn.(net.Conn)
	if ok {
		mubto.connectionInfo = fmt.Sprintf("MUBTO Connection Info(Local: '%s:%s' RemoteAddress: '%s:%s')",
			conn.LocalAddr().Network(),
			conn.LocalAddr().String(),
			conn.RemoteAddr().Network(),
			conn.RemoteAddr().String())
	} else {
		mubto.connectionInfo = fmt.Sprintf("%+v", mubto.conn)
	}
}

func (mubto *mubtoConnection) GetConnectionInfo() string {
	return mubto.connectionInfo
}

func (mubto *mubtoConnection) Closed() bool {
	return nil == mubto.conn
}

// IsCrawler returns a boolean indicating whether the byte array is from a crawler
// based on a heuristic looking for patterns
func IsCrawler(bytes []byte) bool {
	// bytes will be of the form
	/*
		GET /robots.txt HTTP/1.1
		Host: hostname
		Connection: Keep-Alive
		User-Agent: Mozilla/5.0 (compatible; Crawler/1.0; +http://crawler.dne/bots)
		From: crawler@probably.ru
		Accept: */ /*
	 */
	return len(bytes) > 10 && strings.HasPrefix(string(bytes), "GET /robots.txt HTTP/1.1")
}

func (mubto *mubtoConnection) Read(mu MarshalUnmarshal) errors.TracerError {
	if mubto.Closed() {
		return NewNetworkError(NewReadOnClosedError(), mubto.GetConnectionInfo())
	}
	mubto.conn.SetReadDeadline(time.Now().Add(mubto.readTimeout))
	var bytes []byte
	var err error
	if mubto.prefixRead {
		bytes, err = LittleEndian.PrefixRead16(mubto.conn)
	} else {
		bytes, err = Read(mubto.conn)
	}
	if io.EOF == err {
		// make sure we don't use the connection again.
		mubto.Close()
		if len(bytes) == 0 {
			return NewNetworkError(NewErrDisconnectNoData(), mubto.GetConnectionInfo())
		}
	} else if nil != err {
		mubto.Close()
		return NewNetworkError(err, mubto.GetConnectionInfo())
	}
	if len(bytes) == 0 || IsCrawler(bytes) {
		return NewNetworkError(NewErrHealthCheck(), mubto.GetConnectionInfo())
	}
	if err := mu.UnmarshalBinary(bytes); nil != err {
		return NewMarshalError(err)
	}
	return nil
}

func (mubto *mubtoConnection) Write(mu MarshalUnmarshal) errors.TracerError {
	if nil == mubto.conn {
		return NewNetworkError(NewWriteOnClosedError(), mubto.GetConnectionInfo())
	}
	data, err := mu.MarshalBinary()
	if nil != err {
		return NewMarshalError(err)
	}
	mubto.conn.SetWriteDeadline(time.Now().Add(mubto.writeTimeout))
	if err := Write(mubto.conn, data); nil != err {
		mubto.Close()
		return NewNetworkError(err, mubto.GetConnectionInfo())
	}
	return nil
}

func (mubto *mubtoConnection) Close() errors.TracerError {
	var err error
	if nil != mubto.conn {
		err = mubto.conn.Close()
		mubto.conn = nil
	}
	return NewNetworkError(err, mubto.GetConnectionInfo())
}
