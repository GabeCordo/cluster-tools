package database

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/database"
	"github.com/GabeCordo/cluster-tools/core/interfaces"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"testing"
)

func generateDatabaseThread(in chan common.ThreadRequest, out chan common.ThreadResponse) *Thread {

	irc := make(chan common.InterruptEvent, 1)
	min := make(chan common.ThreadRequest, 1)
	mout := make(chan common.ThreadResponse, 1)

	cfg := &Config{Debug: true, Timeout: 2.0}
	logger, _ := logging.NewLogger("database")
	thread, _ := New(cfg, logger, "/test/path", "/test/path2",
		irc, in, out, min, mout, in, out, in, out)

	return thread
}

func TestThread_DatabaseStore_ClusterConfig(t *testing.T) {

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterConfig := interfaces.Config{}

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Type:   common.ConfigRecord,
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

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Type:   common.ConfigRecord,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if response.Success {
		t.Error("expected StoreTypeMismatch due to bad store value")
	}

	if !errors.Is(response.Error, StoreTypeMismatch) {
		t.Error("expected StoreTypeMismatch due to bad store value ")
	}
}

func TestThread_DatabaseStore_SupervisorStatistic(t *testing.T) {

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterStatistic := &interfaces.Statistics{}

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Type:   common.StatisticRecord,
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

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterStatistic := &interfaces.Statistics{}

	request := common.ThreadRequest{
		Action: common.CreateAction,
		Type:   common.ConfigRecord,
		Data:   clusterStatistic,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if response.Success {
		t.Error("expected StoreTypeMismatch due to bad store value")
	}

	if !errors.Is(response.Error, StoreTypeMismatch) {
		t.Error("expected StoreTypeMismatch due to bad store value ")
	}
}

func TestThread_DatabaseFetch_ClusterConfig(t *testing.T) {

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterConfig := interfaces.Config{Identifier: "test_cluster"}

	m := "test_module"
	c := "test_cluster"

	in <- common.ThreadRequest{
		Action:      common.CreateAction,
		Type:        common.ConfigRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterConfig,
		Nonce:       1,
	}
	<-out

	request := common.ThreadRequest{
		Action:      common.GetAction,
		Type:        common.ConfigRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedClusterConfigs, ok := (response.Data).([]interfaces.Config)
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

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterStat := &interfaces.Statistics{}
	clusterStat.Threads.NumProvisionedExtractRoutines = 5

	m := "test_module"
	c := "test_cluster"

	in <- common.ThreadRequest{
		Action:      common.CreateAction,
		Type:        common.StatisticRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterStat,
		Nonce:       1,
	}
	<-out

	request := common.ThreadRequest{
		Action:      common.GetAction,
		Type:        common.StatisticRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedClusterStats, ok := (response.Data).([]database.Statistic)
	if !ok {
		t.Error("expected fetched record to be of type []database.Statistic")
		return
	}

	if len(fetchedClusterStats) != 1 {
		t.Error("expected 1 record to be returned")
		return
	}

	if fetchedClusterStats[0].Stats.Threads.NumProvisionedTransformRoutes != clusterStat.Threads.NumProvisionedTransformRoutes {
		t.Error("fetched wrong *cluster.Statistic record")
	}
}

func TestThread_DatabaseDelete_ClusterConfig(t *testing.T) {

	// TODO - fix
	return

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	m := "test_module"
	c := "test_cluster"
	clusterConfig := interfaces.Config{Identifier: c}

	in <- common.ThreadRequest{
		Action:      common.CreateAction,
		Type:        common.ConfigRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterConfig,
		Nonce:       1,
	}
	<-out

	in <- common.ThreadRequest{
		Action:      common.DeleteAction,
		Type:        common.ClusterRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}

	request := common.ThreadRequest{
		Action:      common.GetAction,
		Type:        common.ConfigRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
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

	in := make(chan common.ThreadRequest, 1)
	out := make(chan common.ThreadResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	m := "test_module"
	c := "test_cluster"
	clusterStat := &interfaces.Statistics{}
	clusterStat.Threads.NumProvisionedLoadRoutines = 5

	in <- common.ThreadRequest{
		Action:      common.CreateAction,
		Type:        common.StatisticRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Data:        clusterStat,
		Nonce:       1,
	}
	<-out

	in <- common.ThreadRequest{
		Action:      common.DeleteAction,
		Type:        common.StatisticRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}

	request := common.ThreadRequest{
		Action:      common.GetAction,
		Type:        common.StatisticRecord,
		Identifiers: common.RequestIdentifiers{Module: m, Cluster: c},
		Nonce:       2,
	}
	in <- request
	response := <-out

	if response.Success {
		t.Error("expected the record to have been deleted")
		return
	}
}
