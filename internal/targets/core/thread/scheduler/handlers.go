package scheduler

import (
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
)

func (t *Thread) handleGetJob(request *thread.Request, response *thread.Response) {

	filter, ok := (request.Data).(database.Filter)
	if !ok {
		response.Success = false
		response.Error = thread.BadRequestType
		return
	}

	jobs := t.useCases.GetJobs(filter)
	response.Data = jobs
	response.Success = len(jobs) > 0
}

func (t *Thread) handleGetQueue(request *thread.Request, response *thread.Response) {

	response.Data = t.useCases.GetQueue()
	response.Success = true
}

func (t *Thread) handleCreateJob(request *thread.Request, response *thread.Response) {

	jb, ok := (request.Data).(job.Job)
	if !ok {
		response.Success = false
		response.Error = thread.BadRequestType
		return
	}

	response.Error = t.useCases.CreateJob(&jb)
	response.Success = response.Error == nil
	t.logger.Printf("created job:%s\n", jb.Identifier)
}

func (t *Thread) handleDeleteJob(request *thread.Request, response *thread.Response) {

	filter, ok := (request.Data).(database.Filter)
	if !ok {
		response.Success = false
		response.Error = thread.BadRequestType
		return
	}

	response.Error = t.useCases.DeleteJob(filter)
	response.Success = response.Error == nil
	t.logger.Printf("deleted job:%s\n", filter.Identifier)
}
