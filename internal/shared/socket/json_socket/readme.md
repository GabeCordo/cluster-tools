# JSON Socket
The sockets use byte-encoded JSON to transmit and receive data over a TCP connection.
The json_socket.Client and json_socket.Server implement the socket.Client and socket.Server interfaces.

## Server Architecture
TODO

## Client Architecture

The client is in the _disconnected_ state by default. Attempts to send on the client
will be blocking until connection is established with the server.

The client transitions to the _connecting_ state when we begin establishing a TCP connection with the server.

The client transitions to the _connected_ state when a TCP connection is established with the server.
The client shall begin sending and receiving messages over the TCP connection when in the _connected_ state.

The client transitions to the _disconnecting_ state when:
- the client is unable to send data over the TCP connection;
- the client receives an EOF from the server;
- the client receives and explicit request to disconnect the session.

While in the disconnected state all attempts to send data over the TCP connection shall block. Any pending
messages shall be dropped by the client.