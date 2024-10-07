package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sentmint/cluster-tools/internal/core/database/pipeline"
	"net/http"
)

var client = http.Client{}

func ProvisionRun(processor string, namespaceName string, id uint64, cfg *pipeline.Pipeline, metadata map[string]string) error {

	body := &struct {
		Namespace  string            `json:"namespace"`
		Supervisor uint64            `json:"id"`
		Config     pipeline.Pipeline `json:"pipeline"`
		Metadata   map[string]string `json:"metadata"`
	}{
		namespaceName, id, *cfg, metadata,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	url := fmt.Sprintf("http://%s/run", processor)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.Status != "200 OK" {
		return errors.New("failed to provision new runner")
	}
	return nil
}

func StopRun(processor string, id uint64) error {

	url := fmt.Sprintf("http://%s/run?id=%d", processor, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return errors.New("non 200 error code returned")
	} else {
		return nil
	}
}
