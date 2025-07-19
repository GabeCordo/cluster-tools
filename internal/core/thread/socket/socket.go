package socket

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/thread"
	common "github.com/GabeCordo/Flock/internal/shared/async"
)

func (t *Thread) setupSocketTlsConfig() error {

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
			return err
		}
		t.tls.config = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	return nil
}

func (t *Thread) startNetworkSocket() {

	if t.flags.useTLS && (t.tls.config == nil) {
		panic("tls config cannot be nil")
	}

	ladder := fmt.Sprintf("%s:%d", t.config.Net.Host, t.config.Net.Port)

	var listener net.Listener
	var err error

	if t.flags.useTLS {
		listener, err = tls.Listen("tcp", ladder, t.tls.config)
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
				Pipe:          t.channels.c7,
				ResponseTable: t.responseTables.processor,
				NoncePool:     t.noncePool,
				Timeout:       t.config.Timeout,
			}

			cfg := &processor2.Config{Identifier: id, RemoteAddr: c.RemoteAddr().String()}
			success, err := thread.AddProcessor(mandatory, cfg)
			if !success {
				t.Logger.Alertln("failed to register a new processor on the processor thread")
				err = conn.Close()
				if err != nil {
					t.Logger.Alert(err.Error())
				}
				return
			}

			t.connectionsMux.Lock()
			// add the processor connection to the map of ongoing connections
			if _, found := t.connections[id]; found {
				t.Logger.Alertln("failed to create a local association to the ongoing connection")
				err = conn.Close()
				if err != nil {
					t.Logger.Alert(err.Error())
				}
				t.connectionsMux.Unlock()
				return
			} else {
				t.connections[id] = conn
				t.connectionsMux.Unlock()
			}

			decode := json.NewDecoder(c)

			request := common.Request{}

			for {
				if err = decode.Decode(&request); err != nil {
					log.Println(err)
					break
				} else {
					t.handleNetworkSocketRequest(id, &request)
				}
			}

			err = c.Close()
			if err != nil {
				t.Logger.Alert(err.Error())
			}

			err = thread.DeleteProcessor(mandatory, cfg)
			if err != nil {
				t.Logger.Alertf("dandling processor %s\n", c.RemoteAddr())
			}

			t.Logger.Printf("closed connection from %s\n", c.RemoteAddr())
		}(conn)
	}
}

func (t *Thread) handleNetworkSocketRequest(processorId uint64, request *common.Request) {

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

					config := new(processor2.ModuleConfig)
					err = json.Unmarshal(b, config)
					if err != nil {
						t.Logger.Warnln("received invalid data for Create Module")
						return
					}

					mandatory := thread.Mandatory{
						Pipe:          t.channels.c7,
						ResponseTable: t.responseTables.processor,
						NoncePool:     t.noncePool,
						Timeout:       t.config.Timeout,
					}

					t.Logger.Printf("received module %s (%s)\n", config.Name, config.Version)

					// TODO: needs processor name
					thread.AsyncAddModule(mandatory, processorId, config)
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
						NoncePool:     t.noncePool,
						Timeout:       t.config.Timeout,
					}

					thread.AsyncUpdateRun(mandatory, instance)
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
