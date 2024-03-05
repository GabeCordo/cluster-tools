package api

import (
	"github.com/GabeCordo/cluster-tools/core/interfaces/module"
	"testing"
)

var moduleCfg = &module.Config{Name: "test", Version: 1.0}

func TestCreateModule(t *testing.T) {

	err := ConnectToCore(host, processorCfg)
	if err != nil {
		t.Error(err)
	}

	err = CreateModule(host, processorCfg, moduleCfg)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteModule(t *testing.T) {

	err := ConnectToCore(host, processorCfg)
	if err != nil {
		t.Error(err)
	}

	err = CreateModule(host, processorCfg, moduleCfg)
	if err != nil {
		t.Error(err)
	}

	err = DeleteModule(host, processorCfg, moduleCfg)
	if err != nil {
		t.Error(err)
	}
}
