package api

import (
	"github.com/GabeCordo/cluster-tools/internal/core/processor"
	"testing"
)

var host = "http://localhost:8137"
var processorCfg = &processor.Config{Host: "localhost", Port: 5023}

func TestConnectToCore(t *testing.T) {

	err := ConnectToCore(host, processorCfg)
	if err != nil {
		t.Error(err)
	}
}

func TestDisconnectFromCore(t *testing.T) {

	err := ConnectToCore(host, processorCfg)
	if err != nil {
		t.Error(err)
	}

	err = DisconnectFromCore(host, processorCfg)
	if err != nil {
		t.Error(err)
	}
}
