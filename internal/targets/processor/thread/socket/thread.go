package socket

import (
	"github.com/FortifiedCode/flock/internal/targets/processor/thread"
)

func (t *Thread) Setup() {

	err := t.useCases.SetupSocketTlsConfig()
	if err != nil {
		panic(err)
	}

	events := &Events{thread: t}
	t.useCases.SetupSocketEventHandlers(events)
}

func (t *Thread) Start() {

	go t.useCases.ConnectToServer(*t.Config.Core)

	var iReq *thread.SocketRequest
	var iRsp *thread.ProvisionerResponse
	stop := false

	for {
		select {
		case iReq = <-t.channels.C0:
			{
				t.requestWg.Add(1)
				t.handleRequest(iReq)
				t.requestWg.Done()
			}
		case iRsp = <-t.channels.C2:
			{
				t.handleResponse(iRsp)
			}
		case <-t.channels.close:
			{
				stop = true
			}
		}

		if stop {
			break
		}
	}
}

func (t *Thread) handleRequest(request *thread.SocketRequest) {

	// Edge Case: it is possible that the core drops while the processor is alive
	// Behaviour: the processor shall ignore requests received on its socket until the connection
	// 			  is re-established to the core.
	// TODO : could we buffer messages until the core comes online?
	if !t.useCases.IsConnectedToCore() {
		// Since the socket does not return a response, we cannot do
		// anything until this is enhanced.
		return
	}

	switch request.Action {
	case thread.SocketModuleAdd:
		{
			t.handleSocketModuleAdd(request)
		}
	case thread.SocketRunUpdate:
		{
			t.handleSocketRunUpdate(request)
		}
	case thread.SocketLogAdd:
		{
			t.logger.Warnln("socket log add not implemented")
		}
	default:
		{
			t.logger.Warnln("received invalid socket request action")
		}
	}
}

func (t *Thread) handleResponse(response *thread.ProvisionerResponse) {

	if response == nil {
		t.logger.Warnln("received nil response")
		return
	}

	if response.Error != nil {
		t.logger.Warnf("received error from provisioner %s\n", response.Error)
	}
}

func (t *Thread) Teardown() {

	t.useCases.DisconnectFromCore()
	t.requestWg.Wait()
}
