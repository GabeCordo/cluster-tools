package socket

import (
	"encoding/json"
	"github.com/FortifiedCode/flock/internal/shared/socket"
	"github.com/FortifiedCode/flock/internal/targets/core/database/run"
	"github.com/FortifiedCode/flock/internal/targets/processor/thread"
)

type Events struct {
	thread *Thread // non-owning reference
}

func (events Events) OnConnectEvent() {

	events.thread.logger.Println("requesting the processor to send modules to the core")
	thread.AsyncRequestModules(thread.ProvisionerMandatory{
		Pipe:      events.thread.channels.C1,
		NoncePool: events.thread.noncePool,
	})
}

func (events Events) OnDisconnectEvent() {

	// DO NOTHING
}

func (events Events) OnMessageEvent(request *socket.Message) {

	switch request.Action {
	case socket.Create:
		{
			switch request.Record {
			case socket.Run:
				{
					events.eventCreateRun(request)
				}
			default:
				{
					events.thread.logger.Warnln("received invalid socket create record")
				}
			}
		}
	case socket.Delete:
		{
			switch request.Record {
			case socket.Run:
				{
					events.eventCreateRun(request)
				}
			default:
				{
					events.thread.logger.Warnln("received invalid socket delete record")
				}
			}
		}
	default:
		{
			events.thread.logger.Warnln("received invalid socket request action")
		}
	}
}

func (events Events) eventCreateRun(request *socket.Message) {

	b, err := json.Marshal(request.Data)
	if err != nil {
		events.thread.logger.Warnln("failed to marshal the received data")
		return
	}

	runRequest := new(run.Request)
	err = json.Unmarshal(b, runRequest)
	if err != nil {
		events.thread.logger.Warnln("received invalid data for update run")
		return
	}

	mandatory := thread.ProvisionerMandatory{
		Pipe:      events.thread.channels.C1,
		NoncePool: events.thread.noncePool,
	}
	thread.AsyncRunStart(mandatory, runRequest.Namespace,
		runRequest.Id, runRequest.Config, runRequest.Metadata)
}

func (events Events) eventDeleteRun(request *socket.Message) {

	id, ok := request.Data.(float64)
	if !ok {
		events.thread.logger.Warnln("run delete received value other than uint64")
		return
	}

	mandatory := thread.ProvisionerMandatory{
		Pipe:      events.thread.channels.C1,
		NoncePool: events.thread.noncePool,
	}
	thread.AsyncRunStop(mandatory, uint64(id))
}
