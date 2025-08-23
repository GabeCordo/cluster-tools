package socket

import (
	"errors"
	"fmt"
	"github.com/FortifiedCode/flock/internal/shared/buffers"
	"net"
	"sync"
)

//////////////////////////////////////////////////////////////////////////////
//							Connection Register
//////////////////////////////////////////////////////////////////////////////
//
//	@brief	a connection register is a generic implementation
//			for handling connections received by a server.
//
//			this includes functionality to:
//				1) reserve a unique identifier for each new connection.
//				2) allocate a Connection object for each new connection.
//				3) release a unique identifier for each terminated connection.
//
//////////////////////////////////////////////////////////////////////////////

/* -------------------------- **** Types ***** ---------------------------- */

// ConnectionRegister
// is a container type that holds zero-to-many clients that connect to a server.
//
// Goals:
//   - Handle assigning unique identifiers to new client connection.
//   - Handle creating new client connections.
//
// Who Uses This?:
//   - An implementation of the Server interface to track clients.
type ConnectionRegister struct {
	clientIdPool *buffers.RingBuffer
	connection   struct {
		clients map[ConnectionId]*Connection // maps an identifier to a connection struct
		mutex   sync.RWMutex                 // controls async access to the clients map
		number  uint64                       // the number of clients references held by the handler
	}
	factories struct {
		encoder EncoderFactory // encoder factory is required when creating a *Connection instance
		decoder DecoderFactory // decoder factory is required when creating a *Connection instance
	}
}

/* ------------------------ **** Functions ***** --------------------------- */

// NewConnectionRegister
// is a function that allocates memory for the ConnectionRegister structure.
//
// Thread Safe: Yes,
// Allocates Memory: Yes
func NewConnectionRegister(eF EncoderFactory, dF DecoderFactory) *ConnectionRegister {

	handler := new(ConnectionRegister)
	if handler == nil {
		panic("failed to allocate memory for *ConnectionRegister")
	}

	var err error
	handler.clientIdPool, err = buffers.NewRingBuffer(MaximumNumOfClientsOnServer)
	if err != nil {
		panic(err)
	}

	var i uint64
	for i = 1; i < MaximumNumOfClientsOnServer+1; i++ {
		err = handler.clientIdPool.Add(i)
		if err != nil {
			panic(err)
		}
	}

	handler.factories.encoder = eF
	handler.factories.decoder = dF

	handler.connection.clients = make(map[ConnectionId]*Connection, MaximumNumOfClientsOnServer)
	handler.connection.number = 0

	return handler
}

// GetConnection
// is a function that returns a *Connection pointer for a ConnectionId.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (handler *ConnectionRegister) GetConnection(id ConnectionId) (client *Connection, err error) {

	handler.connection.mutex.RLock()
	defer handler.connection.mutex.RUnlock()

	var found bool
	client, found = handler.connection.clients[id]
	if !found {
		err = errors.New(fmt.Sprintf("Connection %v does not exist", id))
	}

	return client, err
}

// RegisterConnection
// is a function that generates a unique identifier to represent the client and allocated
// memory for a *Connection structure that represents the client.
//
// Thread Safe: Yes,
// Allocates Memory: Yes
func (handler *ConnectionRegister) RegisterConnection(conn net.Conn) (connection *Connection, err error) {

	// lock the connectionMutex to increment the counter in a thread-safe way
	handler.connection.mutex.Lock()
	defer handler.connection.mutex.Unlock() // unlock the connectionMutex now that the counter is acquired

	if handler.connection.number >= MaximumNumOfClientsOnServer {
		// TODO: log that the maximum number of connections was breached!
		err = errors.New("maximum client connections reached")
		return nil, err
	}

	// todo : name this better
	// allocate a new client id
	value, err := handler.clientIdPool.Remove()
	if err != nil {
		return nil, err
	}

	id, ok := (value).(uint64)
	if !ok {
		return nil, errors.New("invalid connection id")
	}

	handler.connection.number++ // todo: handle what happens when the counter overlaps

	// add the processor connection to the map of ongoing connections
	if _, found := handler.connection.clients[ConnectionId(id)]; found {
		err = errors.New("connection with the same id already exists")
		return nil, err
	}

	connection = NewConnection(handler.factories.encoder, handler.factories.decoder)
	connection.Id = ConnectionId(id)
	connection.SetNetConn(conn)
	handler.connection.clients[ConnectionId(id)] = connection

	return connection, err
}

// ReleaseConnection
// is a function that deallocates the *Connection structure and releases the unique
// identifier so that it may be used by other clients.
//
// Thread Safe: Yes,
// Allocates Memory: No
func (handler *ConnectionRegister) ReleaseConnection(id ConnectionId) {

	handler.connection.mutex.Lock()
	defer handler.connection.mutex.Unlock()

	// release the client id back into the pool of usable client ids
	err := handler.clientIdPool.Add(uint64(id))
	if err != nil {
		fmt.Println(err)
	}

	handler.connection.number--
	delete(handler.connection.clients, id)
}
