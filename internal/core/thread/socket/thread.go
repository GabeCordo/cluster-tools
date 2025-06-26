package socket

import (
	"encoding/json"
	"net"

	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/thread"
	common "github.com/GabeCordo/Flock/internal/shared/async"
)

func (t *Thread) Setup() {

	t.accepting = true

	err := t.setupSocketTlsConfig()
	if err != nil {
		panic(err)
	}
}

func (t *Thread) Start() {

	go t.startNetworkSocket()

	var iReq *thread.Request
	var iRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c9:
			{
				oRsp := t.HandleRequest(iReq)
				if oRsp != nil {
					t.channels.c10 <- oRsp
				}
			}
		case iRsp = <-t.channels.c8:
			{
				t.responseTables.processor.Write(iRsp.Nonce, iRsp)
			}
		}
	}
}

func (t *Thread) HandleRequest(request *thread.Request) (response *thread.Response) {

	response = thread.NewResponse(thread.Socket)
	thread.CopyMetadata(request, response)

	switch request.Action {
	case thread.CreateAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					if request.Identifiers.Processor == 0 {
						t.Logger.Warnln("processor identifier missing for create run")
						response.Error = thread.BadRequestType
						return response
					}

					// TODO: any better way to clean this up + stop using strings for lookup
					t.mutex.RLock()
					t.connectionsMux.RLock()

					var connection net.Conn
					if c, found := t.connections[request.Identifiers.Processor]; !found {
						t.Logger.Warnf("no processor exists with the identifier %d\n", request.Identifiers.Processor)
						t.mutex.RUnlock()
						response.Error = thread.BadRequestType
						t.mutex.RUnlock()
						t.connectionsMux.RUnlock()
						return response
					} else {
						connection = c
						t.mutex.RUnlock()
						t.connectionsMux.RUnlock()
					}

					runRequest, ok := request.Data.(run.Request)
					if !ok {
						t.Logger.Warnln("create run was not given a run.Request type")
						response.Error = thread.BadRequestType
						return response
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
					response.Data = request.Identifiers.Supervisor
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

					t.connectionsMux.RLock()

					var connection net.Conn
					if c, found := t.connections[request.Identifiers.Processor]; found {
						connection = c
						t.connectionsMux.RUnlock()
					} else {
						t.Logger.Warnln("processor identifier not found for delete run")
						response.Error = thread.InternalError
						t.connectionsMux.RUnlock()
						return response
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

	return response
}

func (t *Thread) HandleResponse(iReq *thread.Request, iRsp *thread.Response) (oRsp *thread.Response) {

	panic(thread.NotImplemented)
}

func (t *Thread) Teardown() {
	t.accepting = false
}
