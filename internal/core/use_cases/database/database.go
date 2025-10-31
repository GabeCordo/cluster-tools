package database

import (
	"errors"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/plover"
)

func (uc UseCases) CreatePipelineRecord(namespaceId, pipelineId string, pipelineData *plover.PipelineIR) (err error) {

	if pipelineData == nil {
		return errors.New("the *plover.PipelineIR cannot be nil")
	}

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
	for i, result := range results {
		pipelines[i] = result
	}

	return pipelines, err
}

func (uc UseCases) GetStatisticRecord(namespaceId, pipelineId string) (statistics []*plover.Statistics, err error) {

	results := uc.StatisticDatabase.Get(database.Filter{
		Namespace: namespaceId,
		Pipeline:  pipelineId,
	})

	statistics = make([]*plover.Statistics, len(results))
	for i, result := range results {
		statistics[i] = result
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

func (uc UseCases) ReplacePipelineRecord(namespaceId, pipelineId string, pipelineData *plover.PipelineIR) (err error) {

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
