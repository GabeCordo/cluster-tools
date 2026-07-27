package socket

import (
	"encoding/json"
	"net"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/socket"
	processor2 "github.com/GabeCordo/FunctionScheduler/internal/targets/core/component/processor"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database/run"
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/thread"
)

type Events struct {
	thread *Thread
}

func (events Events) OnClientConnectEvent(id socket.ConnectionId, conn net.Conn) error {

	mandatory := thread.Mandatory{
		Pipe:          events.thread.channels.c7,
		Log:           events.thread.logger,
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
		Log:           events.thread.logger,
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
						events.thread.logger.Println(err.Error())
						return
					}

					config := new(ScalingFunctions.ModuleIR)
					err = json.Unmarshal(b, config)
					if err != nil {
						events.thread.logger.Printf("ConnectionId %d invalid data for Create Module", id)
						return
					}

					events.thread.logger.Printf(
						"ConnectionId %d received a new module %s:%s\n",
						id,
						config.Identifier,
						config.Version,
					)

					mandatory := thread.Mandatory{
						Pipe:          events.thread.channels.c7,
						Log:           events.thread.logger,
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
						events.thread.logger.Printf("ConnectionId %d to marshal the received data", id)
						return
					}

					instance := new(run.Run)
					err = json.Unmarshal(b, instance)
					if err != nil {
						events.thread.logger.Printf("ConnectionId %d received invalid data for update run", id)
						return
					}

					mandatory := thread.Mandatory{
						Pipe:          events.thread.channels.c7,
						Log:           events.thread.logger,
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
