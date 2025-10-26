package database

import (
	"errors"
	"fmt"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/core/database/pipeline"
	"github.com/FortifiedCode/plover"
)

func (uc UseCases) CreatePipelineRecord(namespaceId, pipelineId string, pipelineData *plover.PipelineIR) (err error) {

	err = plover.CleanupIR(pipelineData)
	_, err = uc.PipelineDatabase.Create(
		database.Filter{
			Namespace:  namespaceId,
			Identifier: pipelineId,
		},
		pipelineData,
	)

	return err
}

func (uc UseCases) CreateStatisticRecord(namespaceId, pipelineId string, statisticData *plover.Statistics) (err error) {

	_, err = uc.StatisticDatabase.Create(
		database.Filter{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		statisticData,
	)
	return err
}

func (uc UseCases) GetPipelineRecord(namespaceId, pipelineId string) (pipelines []*plover.PipelineIR, err error) {

	results := uc.PipelineDatabase.Get(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	})

	pipelines = make([]*plover.PipelineIR, len(results))
	ok := true

	var p *plover.PipelineIR
	for i, result := range results {
		p, ok = result.(*plover.PipelineIR)
		if ok {
			pipelines[i] = p
		} else {
			errStr := fmt.Sprintf("expected type '*pipeline.Data' in index %d", i)
			err = errors.New(errStr)
			break
		}
	}

	return pipelines, err
}

func (uc UseCases) GetStatisticRecord(namespaceId, pipelineId string) (statistics []*plover.Statistics, err error) {

	results := uc.StatisticDatabase.Get(database.Filter{
		Namespace: namespaceId,
		Pipeline:  pipelineId,
	})

	statistics = make([]*plover.Statistics, len(results))

	var s *plover.Statistics
	var ok bool

	for i, result := range results {
		s, ok = result.(*plover.Statistics)
		if ok {
			statistics[i] = s
		} else {
			errStr := fmt.Sprintf("expected type '*statistic.Statistics' in index %d", i)
			err = errors.New(errStr)
			break
		}
	}

	return statistics, err
}

func (uc UseCases) DeletePipelineRecord(namespaceId, pipelineId string) (err error) {

	err = uc.PipelineDatabase.Delete(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	})
	return err
}

func (uc UseCases) DeleteStatisticRecord(namespaceId string) (err error) {

	err = uc.StatisticDatabase.Delete(database.Filter{
		Namespace: namespaceId,
	})
	return err
}

func (uc UseCases) ReplacePipelineRecord(namespaceId, pipelineId string, pipelineData *pipeline.Pipeline) (err error) {

	err = uc.PipelineDatabase.Replace(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	}, pipelineData)
	return err
}

func (uc UseCases) LoadDatabases(folder string) (err error) {

	// todo : should we remove this use case if it's not used?

	return nil
}

func (uc UseCases) SaveDatabases(pipelineFolder, statisticsFolder string) {

	if err := uc.PipelineDatabase.Save(pipelineFolder); err != nil {
		uc.Logger.Alertf("failed to save configs created during runtime %s\n", err.Error())
	} else {
		uc.Logger.Printf("saved pipelines to %s\n", pipelineFolder)
	}

	if err := uc.StatisticDatabase.Save(statisticsFolder); err != nil {
		uc.Logger.Alertf("failed to save statistics created during runtime %s\n", err.Error())
	} else {
		uc.Logger.Printf("saved statistics to %s\n", statisticsFolder)
	}
}
