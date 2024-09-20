package api

import (
	"github.com/GabeCordo/clarence/internal/interfaces"
	"testing"
)

var moduleCfg = &interfaces.ModuleConfig{Name: "test", Version: 1.0}

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
