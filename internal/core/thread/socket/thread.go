package socket

import (
	processor2 "github.com/GabeCordo/Flock/internal/core/component/processor"
	"github.com/GabeCordo/Flock/internal/core/database/run"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) Setup() {

	err := t.useCases.SetupSocketTlsConfig()
	if err != nil {
		panic(err)
	}
}

func (t *Thread) Start() {

	go t.useCases.StartNetworkSocket(t.config.Net.Host, t.config.Net.Port,
		func(pId uint64, pAddr string) error {
			mandatory := thread.Mandatory{
				Pipe:          t.channels.c7,
				ResponseTable: t.responseTables.processor,
				NoncePool:     t.noncePool,
				Timeout:       t.config.Timeout,
			}

			cfg := processor2.Config{Identifier: pId, RemoteAddr: pAddr}
			_, err := thread.AddProcessor(mandatory, &cfg)
			return err
		},
		func(pId uint64, pAddr string) error {
			mandatory := thread.Mandatory{
				Pipe:          t.channels.c7,
				ResponseTable: t.responseTables.processor,
				NoncePool:     t.noncePool,
				Timeout:       t.config.Timeout,
			}

			cfg := processor2.Config{Identifier: pId, RemoteAddr: pAddr}
			err := thread.DeleteProcessor(mandatory, &cfg)
			return err
		},
		func(pId uint64, m *processor2.ModuleConfig) {
			mandatory := thread.Mandatory{
				Pipe:          t.channels.c7,
				ResponseTable: t.responseTables.processor,
				NoncePool:     t.noncePool,
				Timeout:       t.config.Timeout,
			}

			// TODO: needs processor name
			thread.AsyncAddModule(mandatory, pId, m)
		},
		func(r *run.Run) {
			mandatory := thread.Mandatory{
				Pipe:          t.channels.c7,
				ResponseTable: t.responseTables.processor,
				NoncePool:     t.noncePool,
				Timeout:       t.config.Timeout,
			}

			thread.AsyncUpdateRun(mandatory, r)
		},
	)

	var iReq *thread.Request
	var iRsp *thread.Response
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c9:
			{
				oRsp = t.HandleRequest(iReq)
				if oRsp != nil {
					t.channels.c10 <- oRsp
				}
			}
		case iRsp = <-t.channels.c8:
			{
				t.responseTables.processor.Write(iRsp.Nonce, iRsp)
			}
		case <-t.channels.close:
			{
				// shutting down the socket thread
				break
			}
		}
		oRsp = nil
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
					t.handleCreateRun(request, response)
				}
			default:
				{
					t.logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	case thread.DeleteAction:
		{
			switch request.Type {
			case thread.RunRecord:
				{
					t.handleDeleteRun(request, response)
				}
			default:
				{
					t.logger.Warn(thread.UnknownRequest.Error())
					response.Error = thread.BadRequestType
				}
			}
		}
	default:
		{
			t.logger.Warn(thread.UnknownRequest.Error())
			response.Error = thread.BadRequestType
		}
	}

	return response
}

func (t *Thread) HandleResponse(iReq *thread.Request, iRsp *thread.Response) (oRsp *thread.Response) {

	panic(thread.NotImplemented)
}

func (t *Thread) Teardown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
