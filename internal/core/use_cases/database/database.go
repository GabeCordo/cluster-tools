package database

import (
	"errors"
	"fmt"
	"github.com/FortifiedCode/flock/internal/core/database"
	"github.com/FortifiedCode/flock/internal/core/database/statistic"
	"github.com/FortifiedCode/plover"
	"log"
	"time"
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

func (uc UseCases) GetPipelineRecord(namespaceId, pipelineId string) (pipelines []plover.PipelineIR, err error) {

	results := uc.PipelineDatabase.Get(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	})

	configs := make([]plover.PipelineIR, len(results))
	ok := true

	for i, result := range results {
		configs[i], ok = result.(plover.PipelineIR)
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

func (uc UseCases) ReplacePipelineRecord(namespaceId, pipelineId string, pipelineData *plover.PipelineIR) (err error) {

	err = uc.PipelineDatabase.Replace(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	}, pipelineData)
	return err
}

func (uc UseCases) LoadDatabases(folder string) (err error) {

	if err = uc.PipelineDatabase.Load(folder); err != nil {
		log.Panicf("could not load saved configs, statistic 'etl doctor' to verify the configuration is valid %s\n",
			err.Error())
	}

	// some configs may have carried over from previous runs
	// let the operator know these configs are being loaded into the
	// flock without having to query the database over HTTP
	uc.PipelineDatabase.Print()

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
