package database

import (
	"errors"
	"github.com/GabeCordo/ScalingFunctions"
	"testing"

	in_memory2 "github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job/in_memory"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/pipeline/in_memory"
	in_memory3 "github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic/in_memory"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/thread"
	database2 "github.com/GabeCordo/DistributedFunctions/internal/targets/core/use_cases/database"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/logging/text_logging"
)

func generateDatabaseThread(in chan *thread.Request, out chan *thread.Response) *Thread {

	irc := make(chan thread.InterruptEvent, 1)
	Min := make(chan *thread.Request, 1)
	Mout := make(chan *thread.Response, 1)

	sD := in_memory3.NewLocalDatabase()
	cD := in_memory.NewLocalPipelineDatabase()
	jD := in_memory2.NewLocalJobDatabase()

	logger, _ := text_logging.New("database")

	useCases := database2.UseCases{
		PipelineDatabase:  cD,
		JobDatabase:       jD,
		StatisticDatabase: sD,
		Logger:            logger,
	}

	cfg := &Config{Debug: true, Timeout: 2.0}

	th, _ := New(cfg, logger, useCases,
		irc, in, out, Min, Mout, in, out, in, out, in, out)

	return th
}

func TestThread_DatabaseStore_ClusterConfig(t *testing.T) {

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	namespaceId := "common"
	pipelineId := "foo"

	p := &ScalingFunctions.PipelineIR{Identifier: pipelineId}

	request := &thread.Request{
		Action: thread.CreateAction,
		Type:   thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		Data:  p,
		Nonce: 1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error)
	}
}

func TestThread_DatabaseStore_ClusterConfig2(t *testing.T) {

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	request := &thread.Request{
		Action: thread.CreateAction,
		Type:   thread.PipelineRecord,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if response.Success {
		t.Error("expected StoreTypeMismatch due to bad database value")
	}

	if !errors.Is(response.Error, StoreTypeMismatch) {
		t.Error("expected StoreTypeMismatch due to bad database value ")
	}
}

func TestThread_DatabaseStore_SupervisorStatistic(t *testing.T) {

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	clusterStatistic := &ScalingFunctions.Statistics{}

	request := thread.Request{
		Action: thread.CreateAction,
		Type:   thread.StatisticRecord,
		Data:   clusterStatistic,
		Nonce:  1,
	}
	in <- &request

	response := <-out

	if !response.Success {
		t.Error(response.Error)
	}
}

func TestThread_DatabaseStore_SupervisorStatistic2(t *testing.T) {

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	clusterStatistic := &ScalingFunctions.Statistics{}

	request := &thread.Request{
		Action: thread.CreateAction,
		Type:   thread.PipelineRecord,
		Data:   clusterStatistic,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if response.Success {
		t.Error("expected StoreTypeMismatch due to bad database value")
	}

	if !errors.Is(response.Error, StoreTypeMismatch) {
		t.Error("expected StoreTypeMismatch due to bad database value ")
	}
}

func TestThread_DatabaseFetch_ClusterConfig(t *testing.T) {

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	namespaceId := "test_namespace"
	pipelineId := "test_pipeline"

	pipelineRecord := &ScalingFunctions.PipelineIR{Identifier: pipelineId}

	in <- &thread.Request{
		Action: thread.CreateAction,
		Type:   thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		Data:  pipelineRecord,
		Nonce: 1,
	}
	<-out

	request := &thread.Request{
		Action: thread.GetAction,
		Type:   thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		Nonce: 2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedPipelines, ok := (response.Data).([]*ScalingFunctions.PipelineIR)
	if !ok {
		t.Error("expected fetched record to be of type []cluster.pipeline")
		return
	}

	if len(fetchedPipelines) != 1 {
		t.Error("expected 1 record to be returned")
		return
	}

	if fetchedPipelines[0].Identifier != pipelineRecord.Identifier {
		t.Error("fetched wrong cluster.pipeline record")
	}
}

func TestThread_DatabaseFetch_SupervisorStatistic(t *testing.T) {

	in := make(chan *thread.Request, 2)
	out := make(chan *thread.Response, 2)

	th := generateDatabaseThread(in, out)
	go th.Start()

	stat := &ScalingFunctions.Statistics{}
	stat.Functions = make([]ScalingFunctions.FunctionStatistic, 3)
	stat.Pipes = make([]ScalingFunctions.PipeStatistic, 2)
	stat.Functions[0].Provisions = 5

	namespaceId := "test_namespace"
	pipelineId := "test_pipeline"

	in <- &thread.Request{
		Action: thread.CreateAction,
		Type:   thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		Data:  stat,
		Nonce: 1,
	}
	<-out

	in <- &thread.Request{
		Action: thread.GetAction,
		Type:   thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{
			Namespace: namespaceId,
			Pipeline:  pipelineId,
		},
		Nonce: 2,
	}

	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedClusterStats, ok := (response.Data).([]*ScalingFunctions.Statistics)
	if !ok {
		t.Error("expected fetched record to be of type []database.Statistic")
		return
	}

	if len(fetchedClusterStats) != 1 {
		t.Error("expected 1 record to be returned")
		return
	}

	if fetchedClusterStats[0].Functions[1].Provisions != stat.Functions[1].Provisions {
		t.Error("fetched wrong *cluster.Statistic record")
	}
}

func TestThread_DatabaseDelete_ClusterConfig(t *testing.T) {

	// TODO - fix
	t.Skip("test case is failing and requires fixes")

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	m := "test_module"
	c := "test_cluster"
	clusterConfig := ScalingFunctions.PipelineIR{Identifier: c}

	in <- &thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Function: c},
		Data:        clusterConfig,
		Nonce:       1,
	}
	<-out

	in <- &thread.Request{
		Action:      thread.DeleteAction,
		Type:        thread.FunctionRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Function: c},
		Nonce:       2,
	}

	request := &thread.Request{
		Action:      thread.GetAction,
		Type:        thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Function: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if response.Success {
		t.Error("expected the record to have been deleted")
		return
	}
}

func TestThread_DatabaseDelete_SupervisorStatistic(t *testing.T) {

	// TODO - must be fixed in future
	t.Skip("test case is failing and requires fixes")

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	m := "test_module"
	c := "test_cluster"
	clusterStat := &ScalingFunctions.Statistics{}
	clusterStat.Functions = make([]ScalingFunctions.FunctionStatistic, 3)
	clusterStat.Pipes = make([]ScalingFunctions.PipeStatistic, 2)
	clusterStat.Functions[2].Provisions = 5

	in <- &thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Function: c},
		Data:        clusterStat,
		Nonce:       1,
	}
	<-out

	in <- &thread.Request{
		Action:      thread.DeleteAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Function: c},
		Nonce:       2,
	}

	request := &thread.Request{
		Action:      thread.GetAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Function: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if response.Success {
		t.Error("expected the record to have been deleted")
		return
	}
}
