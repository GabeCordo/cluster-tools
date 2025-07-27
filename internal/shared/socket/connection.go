package socket

import (
	"errors"
	"log"
	"net"
)

//////////////////////////////////////////////////////////////////////////////
//							  	Connection
//////////////////////////////////////////////////////////////////////////////
//
//	@brief  a connection is a generic implementation for handling
//			send, receive, and interrupt events as an epoll loop.
//
//			a blocking function EPoll listens on 3 channels (in,
//			out, interrupt) handling events synchronously.
//
//			a connection also handles edge cases such as:
//				- reestablishing a severed connection
//				- failures to write to the socket
//				- failures to read from a socket
//
//////////////////////////////////////////////////////////////////////////////

/* ------------------------ **** Constants ***** -------------------------- */

// connectionEventBufferRoom
// is the default size used to create a buffer channel.
const connectionEventBufferRoom = 2

/* -------------------------- **** Enums ***** ---------------------------- */

// connectionEvents
// is a type used to describe an event that occurred during a connection.
type connectionEvents uint8

const (
	eventFunctionFailed connectionEvents = iota // the connection failed to send or received data.
	eventForceClose                             // the connection was explicitly terminated.
)

/* -------------------------- **** Types ***** ---------------------------- */

// Connection
// is a type that handles receives, sending, and closing a golang network connection.
type Connection struct {
	Id        ConnectionId
	Value     net.Conn
	factories struct {
		encoder EncoderFactory
		decoder DecoderFactory
	}
	queues struct {
		interrupts chan connectionEvents
		out        chan *Message
		in         chan *Message
	}
	flags struct {
		connected        bool
		attemptReconnect bool
	}
}

/* ------------------------- **** Functions ***** --------------------------- */

// NewConnection
// is a constructor for the creating a Connection type.
//
// Variables:
// eF EncoderFactory: used to send data on a network connection.
// dF DecoderFactory: used to received data on a network connection.
//
// Returns:
// The *Connection instance.
//
// Allocates Memory: Yes,
// Thread Safe: Yes
func NewConnection(eF EncoderFactory, dF DecoderFactory) *Connection {

	c := new(Connection)

	c.factories.encoder = eF
	c.factories.decoder = dF

	c.flags.attemptReconnect = true
	c.flags.connected = false

	// the 'interrupt' queue shall be non-blocking
	c.queues.interrupts = make(chan connectionEvents, connectionEventBufferRoom)
	// the 'in' and 'out' queues shall be blocking
	c.queues.out = make(chan *Message)
	c.queues.in = make(chan *Message)

	return c
}

// SetNetConn
// links the socket connection to the Connection object.
//
// Thread Safe: No
// Allocates Memory: No
func (c *Connection) SetNetConn(conn net.Conn) {
	c.flags.connected = true
	c.Value = conn
}

// IsConnected
// returns whether the connection is active.
//
// Thread Safe: No,
// Allocates Memory: No
func (c *Connection) IsConnected() bool {
	return c.flags.connected
}

// AttemptReconnect
// returns whether the connection should attempt to reconnect
// to the next endpoint after the connection is detected as down.
//
// Thread Safe: No,
// Allocates Memory: No
func (c *Connection) AttemptReconnect() bool {
	return c.flags.attemptReconnect
}

// EPoll
// a blocking function that listens to asynchronous events to
// receive, send, or interrupt on the connection.
//
// Thread Safe: No,
// Allocates Memory: No
func (c *Connection) EPoll(onMessage func(*Message)) error {

	if c.Value == nil {
		return errors.New("cannot epoll on a nil connection")
	}

	var err error

	// The pointer to the incoming or outgoing message shall be copied
	// to 'data' after being removed from the in/out channel.
	var data *Message

	// The 'stop' flag indicates whether we should break from the for-loop
	// before blocking on the next channel listen.
	//
	// The 'stop' flag is set when:
	//		1. The endpoint receives EOF from the server.
	//		2. The endpoint fails to send data on the socket.
	//		3. The endpoint explicitly closes the socket.
	stop := false

	// The event loop listening to incoming channel messages.
	//
	//		interrupt queue <- an event has been received that implies
	//						   we need to close the socket.
	//		in queue  		<- a message needs to be transmitted to the target endpoint.
	//		out queue 		<- a message has been received from the target endpoint.
	encoder := c.factories.encoder.NewEncoder(c.Value)
	for {
		select {
		case interrupt := <-c.queues.interrupts:
			{
				// setting the flag to 'false' makes other code aware of the TCP socket
				// state before attempting to send data using the client.
				c.flags.connected = false

				// once exiting the for-loop, the default behaviour is to
				// re-establish the TCP connection with the server.
				//
				// edge-case: don't attempt to reconnect if the endpoint is closing
				//			  the connection. The endpoint has likely closed connection
				//			  to terminate the program.
				if interrupt == eventForceClose {
					c.flags.attemptReconnect = false

					// initiate the termination of the TCP connection with the server
					//
					// closing the socket will cause the client.listen() function to
					// received an EOF from the decoder. Upon receiving an EOF, the
					// goroutine will terminate.
					err = c.Value.Close()
					if err != nil {
						log.Printf("Error closing connection: %v", err)
					}
				}

				// stop the event-loop from continuing
				stop = true
			}
		case data = <-c.queues.out:
			{
				err = encoder.Encode(data)
				if err != nil {
					// the interrupt queue allows for buffer messages
					// therefore, messages pushed to the channel will not block
					c.queues.interrupts <- eventFunctionFailed
				}
			}
		case data = <-c.queues.in:
			{
				onMessage(data)
			}
		}

		// stop epoll-loop
		if stop {
			break
		}
	}

	return nil
}

// Listen
// is a blocking function that generates epoll events indicating
// that data was received on the network connection.
//
// The EPoll loop will receive the asynchronous event generated
// by the Listen function to perform some functionality.
//
// Order of Events:
//  1. The Listen() function creates a new encoder for the socket.
//  2. The encoder blocks execution until a Message is received.
//  3. The Listen() function pushes the received Message to a channel
//     representing data received on the socket.
//  4. The EPoll() function pops the data received on the channel and
//     handles the received Message.
//
// Thread Safe: No,
// Allocates Memory: Yes
func (c *Connection) Listen() (err error) {

	if c.Value == nil {
		err = errors.New("cannot listen on a nil connection")
		return err
	}

	// the decoder turns bytes receives on the TCP connection
	// to socket.Message data structures.
	decoder := c.factories.decoder.NewDecoder(c.Value)
	for {
		// allocate memory for a new Message that is passed to the event-handler
		data := new(Message)
		// the Decode() function blocks until data is received on the socket
		err = decoder.Decode(data)
		if err == nil {
			// push the received data to the epoll event loop
			c.queues.in <- data
		} else {
			// likely an EOF, the TCP connection has severed, we need to
			// re-attempt a connection to the core
			c.queues.interrupts <- eventFunctionFailed
			break
		}
	}

	return nil
}

// Send
// is a blocking function that generates an epoll event to send
// data on the connection.
//
// The EPoll loop will receive the asynchronous event generated by
// Send and forward the Message over the network connection.
//
// Order of Events:
//  1. The Send() function validates Message is non-nil.
//  2. The Send() function pushes the received Message to a channel
//     representing data waiting to be sent on the socket.
//  3. The EPoll() function pops the data received on the channel and
//     sends the Message over the network connection.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (c *Connection) Send(message *Message) error {

	if message == nil {
		return errors.New("cannot send a nil message")
	}

	c.queues.out <- message
	return nil
}

// Terminate
// is a non-blocking function that generates an epoll event to
// sever the network connection.
//
// The EPoll loop will receive the asynchronous event generated
// the Terminate and stop the ongoing network connection. All
// messages in send or receive queue will be discarded upon
// severing the network connection.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (c *Connection) Terminate() {

	c.queues.interrupts <- eventForceClose
}
