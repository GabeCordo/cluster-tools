package database

import (
	"errors"
	database2 "github.com/FortifiedCode/flock/internal/core/use_cases/database"
	"github.com/FortifiedCode/plover"
	"testing"

	"github.com/FortifiedCode/flock/internal/core/database/job"
	"github.com/FortifiedCode/flock/internal/core/database/pipeline"
	"github.com/FortifiedCode/flock/internal/core/database/statistic"
	"github.com/FortifiedCode/flock/internal/core/thread"
	"github.com/FortifiedCode/flock/internal/shared/logging/text_logging"
)

func generateDatabaseThread(in chan *thread.Request, out chan *thread.Response) *Thread {

	irc := make(chan thread.InterruptEvent, 1)
	Min := make(chan *thread.Request, 1)
	Mout := make(chan *thread.Response, 1)

	sD := statistic.NewLocalStatisticDatabase()
	cD := pipeline.NewLocalPipelineDatabase()
	jD := job.NewLocalJobDatabase()

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

	clusterConfig := plover.PipelineIR{}

	request := &thread.Request{
		Action: thread.CreateAction,
		Type:   thread.PipelineRecord,
		Data:   clusterConfig,
		Nonce:  1,
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

	clusterStatistic := &statistic.Statistics{}

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

	clusterStatistic := &statistic.Statistics{}

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

	pipelineRecord := plover.PipelineIR{Identifier: "test_pipeline"}

	n := "test_namespace"
	p := "test_pipeline"

	in <- &thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{Namespace: n, Pipeline: p},
		Data:        pipelineRecord,
		Nonce:       1,
	}
	<-out

	request := &thread.Request{
		Action:      thread.GetAction,
		Type:        thread.PipelineRecord,
		Identifiers: thread.RequestIdentifiers{Namespace: n, Pipeline: p},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedPipelines, ok := (response.Data).([]plover.PipelineIR)
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

	in := make(chan *thread.Request, 1)
	out := make(chan *thread.Response, 1)

	th := generateDatabaseThread(in, out)
	go th.Start()

	stat := &statistic.Statistics{}
	stat.Functions = make([]statistic.FunctionStatistic, 3)
	stat.Pipes = make([]statistic.PipeStatistic, 2)
	stat.Functions[0].Provisions = 5

	n := "test_namespace"
	p := "test_pipeline"

	in <- &thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Namespace: n, Pipeline: p},
		Data:        stat,
		Nonce:       1,
	}
	<-out

	request := &thread.Request{
		Action:      thread.GetAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Namespace: n, Pipeline: p},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedClusterStats, ok := (response.Data).([]statistic.Statistics)
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
	clusterConfig := plover.PipelineIR{Identifier: c}

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
	clusterStat := &statistic.Statistics{}
	clusterStat.Functions = make([]statistic.FunctionStatistic, 3)
	clusterStat.Pipes = make([]statistic.PipeStatistic, 2)
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
