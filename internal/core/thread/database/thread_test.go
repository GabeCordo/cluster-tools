package database

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"github.com/GabeCordo/cluster-tools/internal/database/job"
	"github.com/GabeCordo/cluster-tools/internal/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/thread"
	"github.com/GabeCordo/toolchain/logging"
	"testing"
)

func generateDatabaseThread(in chan thread.Request, out chan thread.Response) *Thread {

	irc := make(chan thread.InterruptEvent, 1)
	min := make(chan thread.Request, 1)
	mout := make(chan thread.Response, 1)

	sD := statistic.NewLocalStatisticDatabase()
	cD := config.NewLocalConfigDatabase()
	jD := job.NewLocalJobDatabase()

	cfg := &Config{Debug: true, Timeout: 2.0}
	logger, _ := logging.NewLogger("database")
	th, _ := New(cfg, logger, sD, cD, jD,
		"/test/path", "/test/path2",
		irc, in, out, min, mout, in, out, in, out, in, out)

	return th
}

func TestThread_DatabaseStore_ClusterConfig(t *testing.T) {

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	clusterConfig := config.Config{}

	request := thread.Request{
		Action: thread.CreateAction,
		Type:   thread.ConfigRecord,
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

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	request := thread.Request{
		Action: thread.CreateAction,
		Type:   thread.ConfigRecord,
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

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	clusterStatistic := &statistic.Statistics{}

	request := thread.Request{
		Action: thread.CreateAction,
		Type:   thread.StatisticRecord,
		Data:   clusterStatistic,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if !response.Success {
		t.Error(response.Error)
	}
}

func TestThread_DatabaseStore_SupervisorStatistic2(t *testing.T) {

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	clusterStatistic := &statistic.Statistics{}

	request := thread.Request{
		Action: thread.CreateAction,
		Type:   thread.ConfigRecord,
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

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	clusterConfig := config.Config{Identifier: "test_cluster"}

	m := "test_module"
	c := "test_cluster"

	in <- thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.ConfigRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterConfig,
		Nonce:       1,
	}
	<-out

	request := thread.Request{
		Action:      thread.GetAction,
		Type:        thread.ConfigRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedClusterConfigs, ok := (response.Data).([]config.Config)
	if !ok {
		t.Error("expected fetched record to be of type []cluster.Config")
		return
	}

	if len(fetchedClusterConfigs) != 1 {
		t.Error("expected 1 record to be returned")
		return
	}

	if fetchedClusterConfigs[0].Identifier != clusterConfig.Identifier {
		t.Error("fetched wrong cluster.Config record")
	}
}

func TestThread_DatabaseFetch_SupervisorStatistic(t *testing.T) {

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	clusterStat := &statistic.Statistics{}
	clusterStat.Threads.NumProvisionedExtractRoutines = 5

	m := "test_module"
	c := "test_cluster"

	in <- thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterStat,
		Nonce:       1,
	}
	<-out

	request := thread.Request{
		Action:      thread.GetAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
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

	if fetchedClusterStats[0].Threads.NumProvisionedTransformRoutes != clusterStat.Threads.NumProvisionedTransformRoutes {
		t.Error("fetched wrong *cluster.Statistic record")
	}
}

func TestThread_DatabaseDelete_ClusterConfig(t *testing.T) {

	// TODO - fix
	return

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	m := "test_module"
	c := "test_cluster"
	clusterConfig := config.Config{Identifier: c}

	in <- thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.ConfigRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterConfig,
		Nonce:       1,
	}
	<-out

	in <- thread.Request{
		Action:      thread.DeleteAction,
		Type:        thread.ClusterRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}

	request := thread.Request{
		Action:      thread.GetAction,
		Type:        thread.ConfigRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
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
	return

	in := make(chan thread.Request, 1)
	out := make(chan thread.Response, 1)

	th := generateDatabaseThread(in, out)
	th.accepting = true
	go th.Start()

	m := "test_module"
	c := "test_cluster"
	clusterStat := &statistic.Statistics{}
	clusterStat.Threads.NumProvisionedLoadRoutines = 5

	in <- thread.Request{
		Action:      thread.CreateAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterStat,
		Nonce:       1,
	}
	<-out

	in <- thread.Request{
		Action:      thread.DeleteAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}

	request := thread.Request{
		Action:      thread.GetAction,
		Type:        thread.StatisticRecord,
		Identifiers: thread.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if response.Success {
		t.Error("expected the record to have been deleted")
		return
	}
}
