package net

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/generator"
	"github.com/Kasita-Inc/gadget/log"
)

const (
	// DefaultJRPC2Port for JSON RPC 2
	DefaultJRPC2Port = 44100
)

// JRPC2Request for calling remote procedures via JSON and a JRPC2Client
type JRPC2Request struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"` // array or map
}

// NewJRPC2Request creates a new request with the passed method and params and generates a new ID.
func NewJRPC2Request(method string, params interface{}) JRPC2Request {
	return JRPC2Request{ID: generator.String(8), Method: method, Params: params}
}

// JRPC2Response for calling remote procedures via JSON and a JRPC2Client
type JRPC2Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}

// JRPC2Client for sending JSON RPC 2 requests
type JRPC2Client interface {
	// Send the passed request
	Send(JRPC2Request) (*JRPC2Response, errors.TracerError)
}

type jrpc2Client struct {
	address string
	dial    func(network, address string) (net.Conn, error)
}

// NewJRPC2Client for communicating with the specified host and port.
func NewJRPC2Client(host string, port int) JRPC2Client {
	return &jrpc2Client{
		address: fmt.Sprintf("%s:%d", host, port),
		dial:    net.Dial,
	}
}

func (client *jrpc2Client) Send(request JRPC2Request) (*JRPC2Response, errors.TracerError) {
	log.Debugf("connecting to JRPC2 address '%s'", client.address)
	conn, err := client.dial("tcp", client.address)
	if nil != err {
		return nil, errors.Wrap(err)
	}
	log.Debugf("successfully connected to JRPC2 address '%s'", client.address)
	defer conn.Close()
	data, err := json.Marshal(request)
	if nil != err {
		return nil, errors.Wrap(err)
	}
	_, err = conn.Write(data)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	response := &JRPC2Response{}
	err = json.NewDecoder(conn).Decode(response)
	return response, errors.Wrap(err)
}
