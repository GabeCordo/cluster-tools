package socket

import (
	"encoding/json"
	"github.com/FortifiedCode/flock/internal/shared/socket"
	processor2 "github.com/FortifiedCode/flock/internal/targets/core/component/processor"
	"github.com/FortifiedCode/flock/internal/targets/core/database/run"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
	"github.com/FortifiedCode/plover"
	"log"
	"net"
)

type Events struct {
	thread *Thread
}

func (events Events) OnClientConnectEvent(id socket.ConnectionId, conn net.Conn) error {

	mandatory := thread.Mandatory{
		Pipe:          events.thread.channels.c7,
		ResponseTable: events.thread.responseTables.processor,
		NoncePool:     events.thread.noncePool,
		Timeout:       events.thread.config.Timeout,
	}

	cfg := processor2.Config{Identifier: uint64(id), RemoteAddr: conn.RemoteAddr().String()}
	_, err := thread.AddProcessor(mandatory, &cfg)
	return err
}

func (events Events) OnClientDisconnectEvent(id socket.ConnectionId, conn string) {

	mandatory := thread.Mandatory{
		Pipe:          events.thread.channels.c7,
		ResponseTable: events.thread.responseTables.processor,
		NoncePool:     events.thread.noncePool,
		Timeout:       events.thread.config.Timeout,
	}

	cfg := processor2.Config{Identifier: uint64(id), RemoteAddr: conn}
	err := thread.DeleteProcessor(mandatory, &cfg)
	if err != nil {
		events.thread.logger.Warnln(err.Error())
	}
}

func (events Events) OnMessageEvent(id socket.ConnectionId, request *socket.Message) {

	switch request.Action {
	case socket.Create:
		{
			switch request.Record {
			case socket.Module:
				{
					b, err := json.Marshal(request.Data)
					if err != nil {
						log.Println(err)
						return
					}

					config := new(plover.ModuleIR)
					err = json.Unmarshal(b, config)
					if err != nil {
						log.Println("received invalid data for Create Module")
						return
					}

					log.Printf("received module %s (%s)\n", config.Identifier, config.Version)

					mandatory := thread.Mandatory{
						Pipe:          events.thread.channels.c7,
						ResponseTable: events.thread.responseTables.processor,
						NoncePool:     events.thread.noncePool,
						Timeout:       events.thread.config.Timeout,
					}

					thread.AsyncAddModule(mandatory, uint64(id), config)
				}
			case socket.Log:
				{
					// NOP
				}
			default:
				{
					// NOP
				}
			}
		}
	case socket.Update:
		{
			switch request.Record {
			case socket.Run:
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

					mandatory := thread.Mandatory{
						Pipe:          events.thread.channels.c7,
						ResponseTable: events.thread.responseTables.processor,
						NoncePool:     events.thread.noncePool,
						Timeout:       events.thread.config.Timeout,
					}

					thread.AsyncUpdateRun(mandatory, instance)
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
