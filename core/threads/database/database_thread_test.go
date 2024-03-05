package database

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/core/components/database"
	"github.com/GabeCordo/cluster-tools/core/interfaces/cluster"
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"testing"
)

func generateDatabaseThread(in chan common.DatabaseRequest, out chan common.DatabaseResponse) *Thread {

	irc := make(chan common.InterruptEvent, 1)
	min := make(chan common.MessengerRequest)
	mout := make(chan common.MessengerResponse)

	cfg := &Config{Debug: true, Timeout: 2.0}
	logger, _ := logging.NewLogger("database")
	thread, _ := New(cfg, logger, "/test/path", "/test/path2",
		irc, in, out, min, mout, in, out, in, out)

	return thread
}

func TestThread_DatabaseStore_ClusterConfig(t *testing.T) {

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterConfig := cluster.Config{}

	request := common.DatabaseRequest{
		Action: common.DatabaseStore,
		Type:   common.ClusterConfig,
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

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	request := common.DatabaseRequest{
		Action: common.DatabaseStore,
		Type:   common.ClusterConfig,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if response.Success {
		t.Error("expected StoreTypeMismatch due to bad store value")
	}

	if errors.Is(response.Error, StoreTypeMismatch) {
		t.Error("expected StoreTypeMismatch due to bad store value ")
	}
}

func TestThread_DatabaseStore_SupervisorStatistic(t *testing.T) {

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterStatistic := &cluster.Statistics{}

	request := common.DatabaseRequest{
		Action: common.DatabaseStore,
		Type:   common.SupervisorStatistic,
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

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterStatistic := &cluster.Statistics{}

	request := common.DatabaseRequest{
		Action: common.DatabaseStore,
		Type:   common.SupervisorStatistic,
		Data:   clusterStatistic,
		Nonce:  1,
	}
	in <- request

	response := <-out

	if response.Success {
		t.Error("expected StoreTypeMismatch due to bad store value")
	}

	if errors.Is(response.Error, StoreTypeMismatch) {
		t.Error("expected StoreTypeMismatch due to bad store value ")
	}
}

func TestThread_DatabaseFetch_ClusterConfig(t *testing.T) {

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterConfig := cluster.Config{Identifier: "test_cluster"}

	m := "test_module"
	c := "test_cluster"

	in <- common.DatabaseRequest{
		Action:  common.DatabaseStore,
		Type:    common.ClusterConfig,
		Module:  m,
		Cluster: c,
		Data:    clusterConfig,
		Nonce:   1,
	}
	<-out

	request := common.DatabaseRequest{
		Action:  common.DatabaseFetch,
		Type:    common.ClusterConfig,
		Module:  m,
		Cluster: c,
		Nonce:   2,
	}
	in <- request
	response := <-out

	if !response.Success {
		t.Error("expected successful fetch of record")
		return
	}

	fetchedClusterConfigs, ok := (response.Data).([]cluster.Config)
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

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	clusterStat := &cluster.Statistics{}
	clusterStat.Threads.NumProvisionedExtractRoutines = 5

	m := "test_module"
	c := "test_cluster"

	in <- common.DatabaseRequest{
		Action:  common.DatabaseStore,
		Type:    common.SupervisorStatistic,
		Module:  m,
		Cluster: c,
		Data:    clusterStat,
		Nonce:   1,
	}
	<-out

	request := common.DatabaseRequest{
		Action:  common.DatabaseFetch,
		Type:    common.SupervisorStatistic,
		Module:  m,
		Cluster: c,
		Nonce:   2,
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

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	m := "test_module"
	c := "test_cluster"
	clusterConfig := cluster.Config{Identifier: c}

	in <- common.DatabaseRequest{
		Action:  common.DatabaseStore,
		Type:    common.ClusterConfig,
		Module:  m,
		Cluster: c,
		Data:    clusterConfig,
		Nonce:   1,
	}
	<-out

	in <- common.DatabaseRequest{
		Action:  common.DatabaseDelete,
		Type:    common.ClusterConfig,
		Module:  m,
		Cluster: c,
		Nonce:   2,
	}

	request := common.DatabaseRequest{
		Action:  common.DatabaseFetch,
		Type:    common.ClusterConfig,
		Module:  m,
		Cluster: c,
		Nonce:   2,
	}
	in <- request
	response := <-out

	if response.Success {
		t.Error("expected the record to have been deleted")
		return
	}
}

func TestThread_DatabaseDelete_SupervisorStatistic(t *testing.T) {

	in := make(chan common.DatabaseRequest, 1)
	out := make(chan common.DatabaseResponse, 1)

	thread := generateDatabaseThread(in, out)
	thread.accepting = true
	go thread.Start()

	m := "test_module"
	c := "test_cluster"
	clusterStat := &cluster.Statistics{}
	clusterStat.Threads.NumProvisionedLoadRoutines = 5

	in <- common.DatabaseRequest{
		Action:  common.DatabaseStore,
		Type:    common.SupervisorStatistic,
		Module:  m,
		Cluster: c,
		Data:    clusterStat,
		Nonce:   1,
	}
	<-out

	in <- common.DatabaseRequest{
		Action:  common.DatabaseDelete,
		Type:    common.SupervisorStatistic,
		Module:  m,
		Cluster: c,
		Nonce:   2,
	}

	request := common.DatabaseRequest{
		Action:  common.DatabaseFetch,
		Type:    common.SupervisorStatistic,
		Module:  m,
		Cluster: c,
		Nonce:   2,
	}
	in <- request
	response := <-out

	if response.Success {
		t.Error("expected the record to have been deleted")
		return
	}
}
