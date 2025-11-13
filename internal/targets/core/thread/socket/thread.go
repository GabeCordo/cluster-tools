package socket

import (
	"github.com/FortifiedCode/flock/internal/flags"
	"github.com/FortifiedCode/flock/internal/shared/terminal"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
)

func (t *Thread) Setup() {

	t.logger.SetColour(terminal.Green)

	err := t.useCases.SetupSocketTlsConfig()
	if err != nil {
		panic(err)
	}

	events := Events{thread: t}
	t.useCases.SetupSocketEventHandlers(events)
}

func (t *Thread) Start() {

	go t.useCases.StartNetworkSocket(t.config.Net.Host, t.config.Net.Port)

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
				if flags.DEBUG {
					t.logger.Printf("Received %s ", iRsp.ToString())
				}
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

	if flags.DEBUG {
		t.logger.Printf("Received %s ", request.ToString())
	}
	
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

func (t *Thread) TearDown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
