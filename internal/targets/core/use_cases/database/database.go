package database

import (
	"errors"

	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
)

func (uc UseCases) CreatePipelineRecord(namespaceId, pipelineId string, pipelineData *ScalingFunctions.PipelineIR) (err error) {

	if pipelineData == nil {
		return errors.New("the *ScalingFunctions.PipelineIR cannot be nil")
	}

	err = ScalingFunctions.CleanupIR(pipelineData)
	_, err = uc.PipelineDatabase.Create(
		database.Filter{
			Namespace:  namespaceId,
			Identifier: pipelineId,
		},
		pipelineData,
	)

	return err
}

func (uc UseCases) CreateStatisticRecord(namespaceId, pipelineId string, statisticData *ScalingFunctions.Statistics) (err error) {

	_, err = uc.StatisticDatabase.Create(
		database.Filter{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		statisticData,
	)
	return err
}

func (uc UseCases) GetPipelineRecord(namespaceId, pipelineId string) (pipelines []*ScalingFunctions.PipelineIR, err error) {

	results := uc.PipelineDatabase.Get(database.Filter{
		Namespace:  namespaceId,
		Identifier: pipelineId,
	})

	pipelines = make([]*ScalingFunctions.PipelineIR, len(results))
	for i, result := range results {
		pipelines[i] = result
	}

	return pipelines, err
}

func (uc UseCases) GetStatisticRecord(namespaceId, pipelineId string) (statistics []*ScalingFunctions.Statistics, err error) {

	results := uc.StatisticDatabase.Get(database.Filter{
		Namespace: namespaceId,
		Pipeline:  pipelineId,
	})

	statistics = make([]*ScalingFunctions.Statistics, len(results))
	for i, result := range results {
		statistics[i] = result
	}

	return statistics, err
}

func (uc UseCases) GetNamespaceRecords() (namespaces []string, err error) {

	foundNamespaces := make(map[string]bool)

	pipelineDbNamespaces, err := uc.PipelineDatabase.Distinct(database.Filter{})
	if err != nil {
		return nil, err
	}

	namespaces = make([]string, len(pipelineDbNamespaces))

	for i, pipelineDbNamespace := range pipelineDbNamespaces {
		namespace, ok := pipelineDbNamespace.(string)
		if !ok {
			return nil, errors.New("distinct() should return a list of namespace strings")
		} else {
			namespaces[i] = namespace
			foundNamespaces[namespace] = true
		}
	}

	statisticDbNamespaces, err := uc.StatisticDatabase.Distinct(database.Filter{})
	if err != nil {
		return nil, err
	}

	for _, statisticDbNamespace := range statisticDbNamespaces {
		namespace, ok := statisticDbNamespace.(string)
		if !ok {
			return nil, errors.New("distinct() should return a list of namespace strings")
		} else if found := foundNamespaces[namespace]; !found {
			namespaces = append(namespaces, namespace)
			foundNamespaces[namespace] = true
		}
	}

	return namespaces, err
}

func (uc UseCases) SummaryOfStatistics(namespaceId string) (fields []string, err error) {

	filter := database.Filter{Namespace: namespaceId}

	var rr []any
	rr, err = uc.StatisticDatabase.Distinct(filter)
	if err != nil {
		return fields, err
	}

	var ok bool

	fields = make([]string, len(rr))
	for idx, r := range rr {
		fields[idx], ok = (r).(string)
		if !ok {
			return fields, errors.New("the field is not a string")
		}
	}

	return fields, err
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

func (uc UseCases) ReplacePipelineRecord(namespaceId, pipelineId string, pipelineData *ScalingFunctions.PipelineIR) (err error) {

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
