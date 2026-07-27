package json_socket

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/socket"
)

///////////////////////////////////////////////////////////////
//						Test Stubs
///////////////////////////////////////////////////////////////

type ServerEventsStub struct {
	numOfClientConnects    uint
	numOfClientDisconnects uint
	numOfMessagesReceived  uint

	numExpectedMessagesReceived sync.WaitGroup
	numExpectedConnections      sync.WaitGroup
}

func (e *ServerEventsStub) OnClientConnectEvent(id socket.ConnectionId, conn net.Conn) error {
	e.numOfClientConnects++
	return nil
}

func (e *ServerEventsStub) OnClientDisconnectEvent(id socket.ConnectionId, conn string) {
	e.numOfClientDisconnects++
	e.numExpectedConnections.Done()
}

func (e *ServerEventsStub) OnMessageEvent(id socket.ConnectionId, request *socket.Message) {
	e.numOfMessagesReceived++
	e.numExpectedMessagesReceived.Done()
}

type ClientEventsStub struct {
	numOfConnectEvents    uint
	numOfDisconnectEvents uint
	numOfMessagesReceived uint
}

func (e *ClientEventsStub) OnConnectEvent() {
	e.numOfConnectEvents++
}

func (e *ClientEventsStub) OnDisconnectEvent() {
	e.numOfDisconnectEvents++
}

func (e *ClientEventsStub) OnMessageEvent(request *socket.Message) {
	e.numOfMessagesReceived++
}

///////////////////////////////////////////////////////////////
//						Unit Tests
///////////////////////////////////////////////////////////////

func TestClientSocket_ConnectToServer(t *testing.T) {

	host := "localhost"
	port := 6123

	server := NewServer()
	serverEvents := &ServerEventsStub{}
	serverEvents.numExpectedConnections.Add(1)
	serverEvents.numExpectedMessagesReceived.Add(1)
	server.SetupEvents(serverEvents)

	go func() {
		err := server.Listen(host, port)
		if err != nil {
			t.Error(err)
		}
	}()

	client := NewClient()
	clientEvents := &ClientEventsStub{}
	client.SetupEvents(clientEvents)

	go func() {
		err := client.Connect(fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			t.Error(err)
		}
	}()

	msg := socket.Message{Action: socket.Ping}
	err := client.Send(&msg)
	if err != nil {
		t.Error(err)
		return
	}

	serverEvents.numExpectedMessagesReceived.Wait()

	err = client.Disconnect()
	if err != nil {
		t.Error(err)
	}

	serverEvents.numExpectedConnections.Wait()

	if serverEvents.numOfClientConnects != 1 {
		t.Error("expected the number of client connects to be 1")
	}

	if serverEvents.numOfClientDisconnects != 1 {
		t.Error("expected the number of client disconnects to be 1")
	}

	if serverEvents.numOfMessagesReceived != 1 {
		t.Error("expected the number of messages received to be 1")
	}
}
