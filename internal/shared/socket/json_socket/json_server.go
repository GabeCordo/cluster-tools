package json_socket

import (
	"crypto/tls"
	"fmt"
	"github.com/FortifiedCode/flock/internal/shared/socket"
	"log"
	"net"
)

//////////////////////////////////////////////////////////////////////////////
//							   JSON Server
//////////////////////////////////////////////////////////////////////////////

/* -------------------------- **** Types ***** ---------------------------- */

// Server
// is an implementation of the socket.Server interface for communicating on a socket using JSON.
type Server struct {
	handler *socket.ConnectionRegister // handles client connections accepted by the server.
	events  socket.ServerEvents        // functions called for socket events
	config  *tls.Config                // the TLS config used by the client when enabled.
}

/* ------------------------ **** Functions ***** --------------------------- */

// NewServer
// is a non-blocking function that allocates memory for the *Server structure.
//
// Thread Safe: Yes,
// Allocates Memory: Yes
func NewServer() *Server {

	server := new(Server)
	server.handler = socket.NewConnectionRegister(EncoderFactory{}, DecoderFactory{})
	server.config = nil
	return server
}

// SetupTLS
// is a non-blocking function that switches the Server into TLS mode. When the
// Server is in TLS mode, it shall use the certificate when listening on a socket.
//
// Thread Safe: No
// Allocates Memory: Yes
func (server *Server) SetupTLS(certificates ...tls.Certificate) error {

	// when a certificate is provided the core shall create a non-TLS socket
	if len(certificates) < 1 {
		return nil
	}

	// when a certificate is provided the core shall create a TLS socket
	server.config = &tls.Config{
		Certificates: []tls.Certificate{certificates[0]},
		MinVersion:   tls.VersionTLS12,
	}

	return nil
}

// SetupEvents
// is a non-blocking function that binds a set of callable functions to the Server.
// The Server shall invoke the functions upon hitting the criteria.
//
// Thread Safe: No,
// Allocates Memory: No
func (server *Server) SetupEvents(events socket.ServerEvents) {
	server.events = events
}

// Listen
// is a blocking function that binds to a socket on the device. The Listen function
// creates new goroutines for each client that attempts a connection to the server.
//
// Requirements:
//   - The Listen function may spawn zero to socket.MaximumNumOfClientsOnServer goroutines.
//   - The Listen function shall allocate a *socket.Connection for each connected client.
//
// Thread Safe: Yes,
// Allocates Memory: Yes
func (server *Server) Listen(host string, port int) (err error) {

	ladder := fmt.Sprintf("%s:%d", host, port)

	var listener net.Listener

	// when the TLS config is non-nil the core shall create a TLS socket
	if server.config != nil {
		listener, err = tls.Listen(useTCP, ladder, server.config)
	} else {
		listener, err = net.Listen(useTCP, ladder)
	}

	if err != nil {
		return err
	}

	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(listener)

	var conn net.Conn
	for {
		// the socket shall accept a new connection from a processor.
		conn, err = listener.Accept()
		if err != nil {
			continue
		}

		// the socket shall spawn a new goroutine to handle the new connection.
		go server.clientConnectionHandler(conn)
	}
}

// clientConnectionHandler
// is an internal blocking function that handles a client connection (connect to disconnect).
func (server *Server) clientConnectionHandler(conn net.Conn) {

	client, err := server.handler.RegisterConnection(conn)
	if err != nil {
		log.Println(err)
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
		return
	}

	err = server.events.OnClientConnectEvent(client.Id, conn)
	if err != nil {
		log.Println(err)
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
		return
	}

	go func() {
		err := client.Listen()
		if err != nil {
			log.Println(err)
		}
	}()

	err = client.EPoll(func(message *socket.Message) {
		server.events.OnMessageEvent(client.Id, message)
	})

	if err != nil {
		log.Println(err)
	}

	server.events.OnClientDisconnectEvent(client.Id, conn.RemoteAddr().String())

	err = conn.Close()
	if err != nil {
		log.Println(err)
	}

	server.handler.ReleaseConnection(client.Id)
}

// Send
// is a blocking function call used to send data to a connected client. The function
// returns an error when there is no connected client with the provided socket.ConnectionId.
//
// The function generates an epoll event to send a message towards the server. Once the epoll
// loop for the client connection is able to dequeue the request, the Send function unblocks.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (server *Server) Send(id socket.ConnectionId, message *socket.Message) error {

	client, err := server.handler.GetConnection(id)
	if err != nil {
		return err
	}

	return client.Send(message)
}
