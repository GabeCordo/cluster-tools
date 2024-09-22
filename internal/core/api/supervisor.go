package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"net/http"
)

var client = http.Client{}

func ProvisionSupervisor(processor string, moduleName, clusterName string, supervisor uint64, cfg *pipeline.Pipeline, metadata map[string]string) error {

	body := &struct {
		Module     string            `json:"module"`
		Cluster    string            `json:"cluster"`
		Config     pipeline.Pipeline `json:"pipeline"`
		Supervisor uint64            `json:"id"`
		Metadata   map[string]string `json:"metadata"`
	}{
		moduleName, clusterName, *cfg, supervisor, metadata,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	url := fmt.Sprintf("http://%s/supervisor", processor)
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
		return errors.New("failed to provision new supervisor")
	}
	return nil
}
