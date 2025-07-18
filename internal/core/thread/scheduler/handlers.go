package scheduler

import (
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/job"
	"github.com/GabeCordo/Flock/internal/core/thread"
)

func (t *Thread) handleGetJob(request *thread.Request, response *thread.Response) {

	if filter, ok := (request.Data).(database.Filter); ok {
		response.Data = t.get(filter)
	} else {
		response.Success = false
		response.Error = thread.BadRequestType
	}
}

func (t *Thread) handleGetQueue(request *thread.Request, response *thread.Response) {

	response.Data = t.queue()
	response.Success = true
}

func (t *Thread) handleCreateJob(request *thread.Request, response *thread.Response) {

	if jb, ok := (request.Data).(job.Job); ok {
		response.Error = t.create(&jb)
		response.Success = response.Error == nil
		t.logger.Printf("created job:%s\n", jb.Identifier)
	} else {
		response.Success = false
		response.Error = thread.BadRequestType
	}
}

func (t *Thread) handleDeleteJob(request *thread.Request, response *thread.Response) {

	if filter, ok := (request.Data).(database.Filter); ok {
		response.Error = t.delete(filter)
		response.Success = response.Error == nil
		t.logger.Printf("deleted job:%s\n", filter.Identifier)
	} else {
		response.Success = false
		response.Error = thread.BadResponseType
	}
}
