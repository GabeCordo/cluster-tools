package messenger

import (
	"github.com/GabeCordo/DistributedFunctions/internal/flags"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
)

func (t *Thread) Setup() {

	t.logger.SetColour(terminal.Blue)
}

func (t *Thread) Start() {

	var iReq *thread.Request
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-t.channels.c3:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c4 <- oRsp
				}
			}
		case iReq = <-t.channels.c22:
			{
				oRsp = t.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					t.channels.c23 <- oRsp
				}
			}
		case iReq = <-t.channels.c17:
			{
				_ = t.handleRequest(iReq)
			}
		case <-t.channels.close:
			{
				// shutting down the messenger thread
				break
			}
		}
		oRsp = nil
	}
}

func (t *Thread) handleRequest(request *thread.Request) (response *thread.Response) {

	var err error

	if flags.DEBUG {
		t.logger.Printf("Received %s", request.ToString())
	}

	switch request.Action {
	case thread.GetAction:
		{
			switch request.Type {
			case thread.SmtpRecord:
				{
					// SMTP record get called BUT is not implemented
					err = thread.NotImplemented
				}
			default:
				{
					err = thread.BadRequestType
				}
			}
		}
	case thread.CloseAction:
		{
			err = t.ProcessCloseLogRequest(request)
		}
	default:
		{
			err = t.ProcessConsoleRequest(request)
		}
	}

	response = thread.NewResponse(thread.Messenger)
	response.Error = err
	response.Success = err == nil
	return response
}

func (t *Thread) Teardown() {

	// send a notification to the Start() goroutine to terminate
	t.channels.close <- thread.Shutdown
}
