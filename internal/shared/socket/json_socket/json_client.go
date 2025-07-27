package json_socket

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/GabeCordo/Flock/internal/shared/socket"
	"log"
	"net"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
//							   JSON Client
//////////////////////////////////////////////////////////////////////////////

/* -------------------------- **** Types ***** ---------------------------- */

// Client
// is an implementation of the socket.Client interface for communicating on a socket using JSON.
type Client struct {
	connection *socket.Connection  // handles listening, receiving, and terminating the connection.
	events     socket.ClientEvents // functions called for socket events.
	config     *tls.Config         // the TLS config used by the client when enabled.
}

/* ------------------------ **** Functions ***** --------------------------- */

// NewClient
// is a function that allocates memory for the *Client structure.
//
// Thread Safe: Yes,
// Allocates Memory: Yes
func NewClient() *Client {

	client := new(Client)
	client.config = nil
	client.connection = socket.NewConnection(EncoderFactory{}, DecoderFactory{})

	return client
}

// SetupTLS
// is a function that sets the certificate pool used by the client
// to establish a TLS connection between a client and server.
//
// Thread Safe: No,
// Allocates Memory: Yes
func (client *Client) SetupTLS(certPool *x509.CertPool) error {

	if certPool == nil {
		return errors.New("the *x509.CertPool cannot be nil")
	}

	client.config = &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return nil
}

// SetupEvents
// is a function that registers a set of events to the Client.
//
// Thread Safe: No,
// Allocates Memory: Yes
func (client *Client) SetupEvents(events socket.ClientEvents) {
	client.events = events
}

func (client *Client) dial(host string) (connection net.Conn, err error) {

	for numOfRetries := 0; numOfRetries < socket.MaxNumberOfRetries; numOfRetries++ {

		if client.config == nil {
			connection, err = net.Dial(useTCP, host)
		} else {
			connection, err = tls.Dial(useTCP, host, client.config)
		}

		if err == nil {
			client.connection.SetNetConn(connection)
			client.events.OnConnectEvent()
			break
		}

		if numOfRetries < (socket.MaxNumberOfRetries - 1) {
			time.Sleep(socket.MaxWaitBeforeRetry)
		} else {
			err = socket.MaxNumberOfAttemptsToConnectHit
		}
	}

	return connection, err
}

// Connect
// Attempts to connect to the server until a maximum number of retries is breached.
// Upon connecting to the server the client will continuously listen for epoll events
// to send or receive data over the established client-server connection.
//
// The Connect function is blocking and spawns an additional process to listen to
// incoming messages from the server.
//
// Thread Safe: Yes,
// Allocates Memory: Yes
func (client *Client) Connect(host string) error {

	conn, err := client.dial(host)
	if err != nil {
		// the client could not establish a connection to the core
		// after exhausting all of its attempted retries.
		return err
	}

	// spawn a new goroutine that listens to incoming requests from the server.
	// pass all incoming requests to the event-loop that remains in the Connect()
	// function call.
	go func() {
		err := client.connection.Listen()
		if err != nil {
			log.Printf("Error listening on %s: %s", host, err)
		}
	}()

	// The event loop that handles all messages being
	// received or sent by the client.
	for {
		// epoll is a blocking function call that monitors the
		// 'in', 'out', and 'interrupt' event queues for the client.
		err = client.connection.EPoll(client.events.OnMessageEvent)
		if err != nil {
			log.Printf("Error listening on %s: %s", host, err)
		}

		// when the processor receives a SIGKILL it will call the Disconnect() function
		// that intentionally closes the socket. When the Disconnect() function closes
		// the socket we will hit this point. We should not attempt to re-connect to the
		// core if we initiated the socket close.
		if !client.connection.AttemptReconnect() {
			break
		}

		// condition:
		//		1. a connection was previously established with the server.
		//		2. the TCP connection to the server was severed.
		//		3. the TCP connection was not severed by the client.
		//
		//	we want to keep re-attempting until we reach the server.
		for {
			// TODO: grow the duration as the number of retries increases
			time.Sleep(socket.DurationBeforeRetry)

			// allow the connection to dial for the maximum number of retries.
			conn, err = client.dial(host)
			if err == nil {
				// re-spawn the listener goroutine for received messages.
				client.connection.SetNetConn(conn)
				go func() {
					err := client.connection.Listen()
					if err != nil {
						log.Printf("Error listening on %s: %s", host, err)
					}
				}()
				// goto the top and restart the client epoll loop
				break
			} else {
				// Edge-Case: the processor was unable to re-establish connection to the core.
			}
		}
	}

	// condition: a connection was established to the server and the client
	//			  closed the TCP connection with the server.
	return nil
}

// Send
// a blocking function to a send a socket.Message to a server.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (client *Client) Send(message *socket.Message) error {

	if client.connection == nil {
		return errors.New("cannot send a nil message")
	}

	return client.connection.Send(message)
}

// IsConnected
// returns whether the Client has an active connection to the Server.
//
// Thread Safe: No,
// Allocates Memory: No
func (client *Client) IsConnected() bool {

	if client.connection == nil {
		return false
	}

	return client.connection.IsConnected()
}

// Disconnect
// severe the Client connection with the Server.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (client *Client) Disconnect() error {

	if client.connection == nil {
		return errors.New("cannot disconnect from a connection that is not established")
	}

	client.connection.Terminate()
	return nil
}
