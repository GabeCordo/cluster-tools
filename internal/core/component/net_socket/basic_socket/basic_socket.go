package basic_socket

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	common "github.com/GabeCordo/Flock/internal/shared/async"
	"log"
	"net"
	"sync"
)

type Socket struct {
	connections      map[uint64]net.Conn
	numOfConnections uint64

	config *tls.Config

	connectionsMux sync.RWMutex
	mutex          sync.RWMutex
}

func New() *Socket {

	socket := new(Socket)
	socket.connections = make(map[uint64]net.Conn)
	socket.numOfConnections = 0
	socket.config = nil

	return socket
}

func (s *Socket) Setup(certificates ...tls.Certificate) error {

	// when a certificate is provided the core shall create a non-TLS socket
	if len(certificates) < 1 {
		return nil
	}

	// when a certificate is provided the core shall create a TLS socket
	s.config = &tls.Config{
		Certificates: []tls.Certificate{certificates[0]},
		MinVersion:   tls.VersionTLS12,
	}

	return nil
}

func (s *Socket) Start(host string, port int,
	CreateProcessor func(pId uint64, pAddr string) error,
	DeleteProcessor func(pId uint64, pAddr string) error,
	CreateModule func(pId uint64, m *processor2.ModuleConfig),
	UpdateRun func(r *run.Run)) {

	ladder := fmt.Sprintf("%s:%d", host, port)

	var listener net.Listener
	var err error

	// when the TLS config is non-nil the core shall create a TLS socket
	if s.config != nil {
		listener, err = tls.Listen("tcp", ladder, s.config)
	} else {
		listener, err = net.Listen("tcp", ladder)
	}

	if err != nil {
		panic(err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
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
		go s.connection(conn, CreateProcessor, DeleteProcessor, CreateModule, UpdateRun)
	}
}

func (s *Socket) connection(conn net.Conn,
	createProcessor func(pId uint64, pAddr string) error,
	deleteProcessor func(pId uint64, pAddr string) error,
	createModule func(pId uint64, m *processor2.ModuleConfig),
	updateRun func(r *run.Run)) {

	// lock the mutex to increment the counter in a thread-safe way
	s.mutex.Lock()

	s.numOfConnections++ // todo: handle what happens when the counter overlaps
	id := s.numOfConnections

	// unlock the mutex now that the counter is acquired
	s.mutex.Unlock()

	err := createProcessor(id, conn.RemoteAddr().String())
	if err != nil {
		err = conn.Close()
		return
	}

	s.connectionsMux.Lock()
	// add the processor connection to the map of ongoing connections
	if _, found := s.connections[id]; found {
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
		s.connectionsMux.Unlock()
		return
	}

	s.connections[id] = conn
	s.connectionsMux.Unlock()

	decode := json.NewDecoder(conn)

	request := common.Request{}

	for {
		if err = decode.Decode(&request); err != nil {
			log.Println(err)
			break
		} else {
			s.handle(id, &request, createModule, updateRun)
		}
	}

	err = conn.Close()
	if err != nil {
		log.Println(err)
	}

	err = deleteProcessor(id, conn.RemoteAddr().String())
}

func (s *Socket) handle(processorId uint64, request *common.Request,
	createModule func(pId uint64, m *processor2.ModuleConfig),
	updateRun func(r *run.Run)) {

	switch request.Action {
	case common.Create:
		{
			switch request.Record {
			case common.Module:
				{
					b, err := json.Marshal(request.Data)
					if err != nil {
						log.Println(err)
						return
					}

					config := new(processor2.ModuleConfig)
					err = json.Unmarshal(b, config)
					if err != nil {
						log.Println("received invalid data for Create Module")
						return
					}

					log.Printf("received module %s (%s)\n", config.Name, config.Version)

					createModule(processorId, config)
				}
			case common.Log:
				{
					// NOP
				}
			default:
				{
					// NOP
				}
			}
		}
	case common.Update:
		{
			switch request.Record {
			case common.Run:
				{
					b, err := json.Marshal(request.Data)
					if err != nil {
						log.Println("failed to marshal the received data")
						return
					}

					instance := new(run.Run)
					err = json.Unmarshal(b, instance)
					if err != nil {
						log.Println("received invalid data for update run")
						return
					}

					updateRun(instance)
				}
			default:
				{
					// NOP
				}
			}
		}
	default:
		{
			// NOP
		}
	}
}

func (s *Socket) GetConnection(pId uint64) (connection net.Conn, err error) {

	// TODO: any better way to clean this up + stop using strings for lookup
	s.mutex.RLock()
	s.connectionsMux.RLock()
	defer s.mutex.RUnlock()
	defer s.connectionsMux.RUnlock()

	if c, found := s.connections[pId]; !found {
		output := fmt.Sprintf("no processor exists with the identifier %d\n", pId)
		err = errors.New(output)
	} else {
		connection = c
	}

	return connection, err
}
