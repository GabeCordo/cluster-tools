package messenger

import (
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (th *Thread) Setup() {

}

func (th *Thread) Start() {

	var iReq *thread.Request
	var oRsp *thread.Response

	for {
		select {
		case iReq = <-th.channels.c3:
			{
				oRsp = th.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					th.channels.c4 <- oRsp
				}
			}
		case iReq = <-th.channels.c22:
			{
				oRsp = th.handleRequest(iReq)
				if oRsp != nil {
					thread.CopyMetadata(iReq, oRsp)
					th.channels.c23 <- oRsp
				}
			}
		case iReq = <-th.channels.c17:
			{
				_ = th.handleRequest(iReq)
			}
		case <-th.channels.close:
			{
				// shutting down the messenger thread
				break
			}
		}
		oRsp = nil
	}
}

func (th *Thread) handleRequest(request *thread.Request) (response *thread.Response) {

	var err error

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
			err = th.ProcessCloseLogRequest(request)
		}
	default:
		{
			err = th.ProcessConsoleRequest(request)
		}
	}

	response = thread.NewResponse(thread.Messenger)
	response.Error = err
	response.Success = err == nil
	return response
}

func (th *Thread) Teardown() {

	th.wg.Wait()

	// send a notification to the Start() goroutine to terminate
	th.channels.close <- thread.Shutdown
}
