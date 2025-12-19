package database

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/targets/core/database/contact"
	"github.com/FortifiedCode/flock/internal/targets/core/thread"
	"github.com/FortifiedCode/plover"
)

func (t *Thread) handleCreatePipelineRecord(request *thread.Request, response *thread.Response) {

	configData, ok := (request.Data).(*plover.PipelineIR)
	if !ok {
		response.Success = false
		response.Error = StoreTypeMismatch
		return
	}

	err := t.useCases.CreatePipelineRecord(request.Identifiers.Namespace, request.Identifiers.Pipeline, configData)
	if err != nil {
		response.Success = false
		response.Error = errors.New("failed to create pipeline record")
		return
	}

	// display to the console the new pipeline database
	t.useCases.PipelineDatabase.Print()

	response.Success = true
}

func (t *Thread) handleCreateStatisticRecord(request *thread.Request, response *thread.Response) {

	statisticsData, ok := (request.Data).(*plover.Statistics)
	if !ok {
		response.Success = false
		response.Error = StoreTypeMismatch
		return
	}

	err := t.useCases.CreateStatisticRecord(request.Identifiers.Namespace, request.Identifiers.Pipeline, statisticsData)
	if err != nil {
		response.Success = false
		response.Error = errors.New("failed to create statistic record")
		return
	}

	// display to the console the new statistic database
	t.useCases.StatisticDatabase.Print()

	response.Success = true
}

func (t *Thread) handleGetPipelineRecord(request *thread.Request, response *thread.Response) {

	configs, err := t.useCases.GetPipelineRecord(request.Identifiers.Namespace, request.Identifiers.Pipeline)
	if err != nil {
		response.Success = false
		response.Error = err
		return
	}

	if len(configs) < 1 {
		response.Error = contact.NotFound
		response.Success = false
	}
	response.Data = configs
}

func (t *Thread) handleGetStatisticRecord(request *thread.Request, response *thread.Response) {

	statistics, err := t.useCases.GetStatisticRecord(request.Identifiers.Namespace, request.Identifiers.Pipeline)
	if err != nil {
		response.Success = false
		response.Error = err
		return
	}

	if len(statistics) < 1 {
		response.Error = contact.NotFound
		response.Success = false
	}
	response.Data = statistics
}

func (t *Thread) handleSummaryStatistics(request *thread.Request, response *thread.Response) {

	summary, err := t.useCases.SummaryOfStatistics(request.Identifiers.Namespace)
	if err != nil {
		response.Success = false
		response.Error = err
	} else {
		response.Data = summary
	}
}

func (t *Thread) handleDeletePipelineRecord(request *thread.Request, response *thread.Response) {

	err := t.useCases.DeletePipelineRecord(request.Identifiers.Namespace, request.Identifiers.Pipeline)

	if err == nil {
		t.useCases.PipelineDatabase.Print()
	}

	response.Success = err == nil
	response.Error = err
}

func (t *Thread) handleDeleteStatisticRecord(request *thread.Request, response *thread.Response) {

	err := t.useCases.DeleteStatisticRecord(request.Identifiers.Namespace)

	if err == nil {
		t.useCases.StatisticDatabase.Print()
	}

	response.Success = err == nil
	response.Error = err
}

func (t *Thread) handleUpdatePipelineRecord(request *thread.Request, response *thread.Response) {

	cfg, ok := (request.Data).(*plover.PipelineIR)
	if !ok {
		response.Success = false
		response.Error = thread.BadRequestType
		return
	}

	err := t.useCases.ReplacePipelineRecord(request.Identifiers.Namespace, request.Identifiers.Pipeline, cfg)

	if err == nil {
		t.useCases.PipelineDatabase.Print()
	}

	response.Success = err == nil
	response.Error = err
}
