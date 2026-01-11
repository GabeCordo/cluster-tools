package socket

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
//							Encoder and Decoders
//////////////////////////////////////////////////////////////////////////////

/* -------------------------- **** Types ***** ---------------------------- */

// Encoder
// is a type than turns a golang data structure into another representation.
type Encoder interface {
	Encode(any) error
}

// EncoderFactory
// is a type that returns an Encoder type.
type EncoderFactory interface {
	NewEncoder(w io.Writer) Encoder
}

// Decoder
// is a type that turns an intermediary representation into a golang data structure.
type Decoder interface {
	Decode(any) error
}

// DecoderFactory
// is a type that returns a Decoder type.
type DecoderFactory interface {
	NewDecoder(r io.Reader) Decoder
}

//////////////////////////////////////////////////////////////////////////////
//							 	 Message
//////////////////////////////////////////////////////////////////////////////

/* -------------------------- **** Enums ***** ---------------------------- */

// Action
// is a request for a unit of functionality to be performed.
type Action uint8

const (
	Ping Action = iota
	Create
	Update
	Delete
)

// Type
// is the form of data an Action shall be performed on.
type Type uint8

const (
	Module Type = iota
	Run
	Log
)

/* -------------------------- **** Types ***** ---------------------------- */

// Message
// is the data type sent across a socket connection.
type Message struct {
	Action Action `json:"action"` // What the endpoint should do.
	Record Type   `json:"type"`   // What the structure of data is.
	Data   any    `json:"data"`   // The data sent to the endpoint.
}

//////////////////////////////////////////////////////////////////////////////
//							   	  Server
//////////////////////////////////////////////////////////////////////////////

/* ------------------------ **** Constants ***** -------------------------- */

// MaximumNumOfClientsOnServer
// is the maximum number of active clients a server can support at once.
const MaximumNumOfClientsOnServer uint64 = 100

/* -------------------------- **** Types ***** ---------------------------- */

// ConnectionId
// identifies the ongoing client connection to a server.
//
// The connection id is not a universal identifier for a client.
// The connection id for a client may change when the client
// disconnects and reconnects to a server endpoint.
type ConnectionId uint64

// ServerEvents
// are a set of functions invoked by a Server implementation.
type ServerEvents interface {
	OnClientConnectEvent(id ConnectionId, conn net.Conn) error
	OnClientDisconnectEvent(id ConnectionId, conn string)
	OnMessageEvent(id ConnectionId, request *Message)
}

// Server
// is an interface for a server endpoint.
type Server interface {
	SetupTLS(...tls.Certificate) error
	SetupEvents(events ServerEvents)
	Listen(host string, port int) error
	Send(id ConnectionId, request *Message) error
}

//////////////////////////////////////////////////////////////////////////////
//								  Client
//////////////////////////////////////////////////////////////////////////////

/* ------------------------- **** Errors ***** ---------------------------- */

var MaxNumberOfAttemptsToConnectHit = errors.New("the client socket could not connect to the server in the max number of attempts")

/* ------------------------ **** Constants ***** -------------------------- */

// MaxNumberOfRetries
// is the maximum number of times a client shall attempt to reconnect to a server.
const MaxNumberOfRetries int = 100

// MaxWaitBeforeRetry
// is the maximum amount of time a client shall wait before reconnecting to a server.
const MaxWaitBeforeRetry = 2 * time.Second

// DurationBeforeRetry
// is the duration before the client attempts to reconnect to the server.
const DurationBeforeRetry = 100 * time.Millisecond

/* -------------------------- **** Types ***** ---------------------------- */

// ClientEvents
// is a set of functions invoked by a Client implementation.
type ClientEvents interface {
	OnConnectEvent()
	OnDisconnectEvent()
	OnMessageEvent(request *Message)
}

// Client
// is an interface for a client endpoint.
type Client interface {
	SetupTLS(certPool *x509.CertPool) error
	SetupEvents(events ClientEvents)
	Connect(host string) error
	IsConnected() bool
	Disconnect() error
	Send(request *Message) error
}
