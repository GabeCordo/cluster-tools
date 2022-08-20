package core

import "ETLFramework/statemachine"

var stateMachineInstance *statemachine.StateMachine

func GetStateMachineInstance() *statemachine.StateMachine {
	if stateMachineInstance == nil {
		stateMachineInstance = statemachine.NewStateMachine()
	}

	return stateMachineInstance
}

func (stateMachineThread *StateMachineThread) Setup() {
	s := GetStateMachineInstance() // initialize the state machine memory address
	_ = s
}

func (stateMachineThread *StateMachineThread) Start() {
	go func() {
		for stateMachineThread.accepting {
			request := <-stateMachineThread.C9
			stateMachineThread.ProcessIncomingRequests(request)
		}
	}()

	go func() {
		for stateMachineThread.accepting {
			request := <-stateMachineThread.C11
			stateMachineThread.ProcessIncomingRequests(request)
		}
	}()

	stateMachineThread.wg.Wait()
}

func (stateMachineThread *StateMachineThread) ProcessIncomingRequests(request StateMachineRequest) {

}

func (stateMachineThread *StateMachineThread) ProcessIncomingResponse(response StateMachineResponse) {

}

func (stateMachineThread *StateMachineThread) Teardown() {

}
