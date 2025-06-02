package socket

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	common "github.com/GabeCordo/Flock/internal"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/processor"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"log"
	"net"
	"os"
)

func (t *Thread) Setup() {

	t.accepting = true

	certificatePath := os.Getenv("CTOOLS_TLS_CERT")
	if certificatePath == "" {
		t.Logger.Warnln("CTOOLS_TLS_CERT environment variable not set")
	}

	keyPath := os.Getenv("CTOOLS_TLS_KEY")
	if keyPath == "" {
		t.Logger.Warnln("CTOOLS_TLS_KEY environment variable not set")
	}

	// [requirements]
	// 1. the core(gateway) shall default to an un-encrypted socket when the TLS cert is missing
	// 2. the core(gateway) shall default to an un-encrypted socket when the TLS key is missing
	// 3. the core shall output a warning message when an un-encrypted socket is opened
	if certificatePath == "" || keyPath == "" {
		t.Logger.Alertln("the gateway has defaulted to an unencrypted socket! Do NOT use in production!")
		t.flags.useTLS = false
	} else {
		t.flags.useTLS = true
	}

	if t.flags.useTLS {
		cert, err := tls.LoadX509KeyPair(certificatePath, keyPath)
		if err != nil {
			panic(err)
		}
		t.tls.config = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
}

func (t *Thread) Start() {

	// REQUEST THREADS

	thread.SetupListener(t.channels.c9, t.channels.c10, &t.accepting, &t.wg, thread.Socket, t.Handle)

	// RESPONSE THREADS

	go func() {
		for response := range t.channels.c8 {
			t.responseTables.processor.Write(response.Nonce, response)
		}
	}()

	// TLS SOCKET

	if t.flags.useTLS && (t.tls.config == nil) {
		panic("tls config cannot be nil")
	}

	laddr := fmt.Sprintf("%s:%d", t.config.Net.Host, t.config.Net.Port)

	var listener net.Listener
	var err error

	if t.flags.useTLS {
		listener, err = tls.Listen("tcp", laddr, t.tls.config)
	} else {
		listener, err = net.Listen("tcp", laddr)
	}

	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		t.Logger.Printf("accepted new connection from %s\n", conn.RemoteAddr())

		go func(c net.Conn) {

			// lock the mutex to increment the counter in a thread-safe way
			t.mutex.Lock()

			t.numOfConnections++ // todo: handle what happens when the counter overlaps
			id := t.numOfConnections

			// unlock the mutex now that the counter is acquired
			t.mutex.Unlock()

			mandatory := thread.Mandatory{
				t.channels.c7,
				t.responseTables.processor,
				t.config.Timeout,
			}

			cfg := &processor.Config{Identifier: id, RemoteAddr: c.RemoteAddr().String()}
			success, err := thread.AddProcessor(mandatory, cfg)
			if !success {
				t.Logger.Alertln("failed to register a new processor on the processor thread")
				conn.Close()
				return
			}

			// add the processor connection to the map of ongoing connections
			if _, found := t.connections[id]; found {
				t.Logger.Alertln("failed to create a local association to the ongoing connection")
				conn.Close()
				return
			} else {
				t.connections[id] = conn
			}

			decode := json.NewDecoder(c)

			request := common.Request{}

			for {
				if err = decode.Decode(&request); err != nil {
					log.Println(err)
					break
				} else {
					t.HandleSocketRequest(id, &request)
				}
			}

			c.Close()

			err = thread.DeleteProcessor(mandatory, cfg)
			if err != nil {
				t.Logger.Alertf("dandling processor %s\n", c.RemoteAddr())
			}

			t.Logger.Printf("closed connection from %s\n", c.RemoteAddr())
		}(conn)
	}
}

func (t *Thread) HandleSocketRequest(processorId uint64, request *common.Request) {

	switch request.Action {
	case common.Create:
		{
			switch request.Record {
			case common.Module:
				{
					b, err := json.Marshal(request.Data)
					if err != nil {
						t.Logger.Warnln("failed to marshal the received data")
						return
					}

					config := new(processor.ModuleConfig)
					err = json.Unmarshal(b, config)
					if err != nil {
						t.Logger.Warnln("received invalid data for Create Module")
						return
					}

					mandatory := thread.Mandatory{
						Pipe:          t.channels.c7,
						ResponseTable: t.responseTables.processor,
						Timeout:       t.config.Timeout,
					}

					t.Logger.Printf("received module %s (%s)\n", config.Name, config.Version)

					// TODO: needs processor name
					_, err = thread.AddModule(mandatory, processorId, config)
					if err != nil {
						log.Println(err)
					}
				}
			case common.Log:
				{
					t.Logger.Alertln("unimplemented log action called")
				}
			default:
				{
					t.Logger.Warnln("received unknown create record from processor")
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
						t.Logger.Warnln("failed to marshal the received data")
						return
					}

					instance := new(run.Run)
					err = json.Unmarshal(b, instance)
					if err != nil {
						t.Logger.Warnln("received invalid data for update run")
						return
					}

					mandatory := thread.Mandatory{
						Pipe:          t.channels.c7,
						ResponseTable: t.responseTables.processor,
						Timeout:       t.config.Timeout,
					}

					if err := thread.UpdateRun(mandatory, instance); err != nil {
						t.Logger.Warnf("failed to update run")
					}
				}
			default:
				{
					t.Logger.Warnln("received unknown update record from processor")
				}
			}
		}
	default:
		{
			t.Logger.Warnln("received invalid action from processor")
		}
	}
}

func (t *Thread) Handle(request *thread.Request, response *thread.Response) {

	switch request.Action {
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{

					if request.Identifiers.Processor == 0 {
						t.Logger.Warnln("processor identifier missing for create run")
						response.Error = thread.BadRequestType
						return
					}

					// TODO: any better way to clean this up + stop using strings for lookup
					t.mutex.RLock()
					var connection net.Conn
					if c, found := t.connections[request.Identifiers.Processor]; !found {
						t.Logger.Warnf("no processor exists with the identifier %s\n", request.Identifiers.Processor)
						t.mutex.RUnlock()
						response.Error = thread.BadRequestType
						return
					} else {
						connection = c
					}
					t.mutex.RUnlock()

					runRequest, ok := request.Data.(run.Request)
					if !ok {
						t.Logger.Warnln("create run was not given a run.Request type")
						response.Error = thread.BadRequestType
						return
					}

					encoder := json.NewEncoder(connection)

					r := common.Request{
						Action: common.Create,
						Record: common.Run,
						Data:   runRequest,
					}
					err := encoder.Encode(r)
					if err != nil {
						t.Logger.Warnln("failed to encode run request")
						response.Error = thread.InternalError
					}
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					if request.Identifiers.Processor == 0 {
						t.Logger.Warnln("missing processor identifier for delete run")
						response.Error = thread.InternalError
					}

					if request.Identifiers.Supervisor == 0 {
						t.Logger.Warnln("missing run id for delete run")
						response.Error = thread.InternalError
					}

					r := common.Request{
						Action: common.Delete,
						Record: common.Run,
						Data:   request.Identifiers.Supervisor,
					}

					var connection net.Conn
					if c, found := t.connections[request.Identifiers.Processor]; found {
						connection = c
					} else {
						t.Logger.Warnln("processor identifier not found for delete run")
						response.Error = thread.InternalError
						return
					}

					encoder := json.NewEncoder(connection)
					err := encoder.Encode(r)
					if err != nil {
						t.Logger.Warnln("failed to encode delete run request")
					}
				}
			default:
				{
					t.Logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	default:
		{
			t.Logger.Warn(thread.UnknownRequest.Error())
			response.Error = thread.BadRequestType
		}
	}
}

func (t *Thread) Teardown() {
	t.accepting = false
}
