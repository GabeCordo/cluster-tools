package database

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/Flock/internal/core/database"
	"github.com/GabeCordo/Flock/internal/core/database/pipeline"
	"github.com/GabeCordo/Flock/internal/core/database/statistic"
	"time"
)

type UseCases struct {
	PipelineDatabase  database.Database
	StatisticDatabase database.Database
	JobDatabase       database.Database
}

func (uc UseCases) CreatePipelineRecord(namespaceId, pipelineId string, pipelineData *pipeline.Pipeline) (err error) {

	_, err = uc.PipelineDatabase.Create(
		database.Filter{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		pipelineData,
	)
	return err
}

func (uc UseCases) CreateStatisticRecord(namespaceId, pipelineId string, statisticData *statistic.Statistics) (err error) {

	_, err = uc.StatisticDatabase.Create(
		database.Filter{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		statistic.Wrapper{ // TODO : depreciate or fix elapsed time
			Timestamp: time.Now(),
			Stats:     *statisticData, // copy TODO: maybe fix this
		},
	)
	return err
}

func (uc UseCases) GetPipelineRecord(namespaceId, pipelineId string) (pipelines []pipeline.Pipeline, err error) {

	results := uc.PipelineDatabase.Get(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	})

	configs := make([]pipeline.Pipeline, len(results))
	ok := true

	for i, result := range results {
		configs[i], ok = result.(pipeline.Pipeline)
		if !ok {
			errStr := fmt.Sprintf("expected type 'pipeline.Pipeline' in index %d", i)
			err = errors.New(errStr)
			break
		}
	}

	return configs, err
}

func (uc UseCases) GetStatisticRecord(namespaceId, pipelineId string) (statistics []statistic.Statistics, err error) {

	results := uc.StatisticDatabase.Get(database.Filter{
		Namespace: namespaceId,
		Pipeline:  pipelineId,
	})

	statistics = make([]statistic.Statistics, len(results))
	ok := true

	for i, result := range results {
		statistics[i], ok = result.(statistic.Statistics)
		if !ok {
			errStr := fmt.Sprintf("expected type 'statistic.Statistics' in index %d", i)
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
		Namespace: namespaceId,
		Pipeline:  pipelineId,
	}, pipelineData)
	return err
}
