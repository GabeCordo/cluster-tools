package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/mango/core/components/supervisor"
	"github.com/GabeCordo/mango/core/interfaces/cluster"
	"net/http"
)

func Log(host string, message string) error {
	return nil
}

func LogWarn(host string, message string) error {
	return nil
}

func LogError(host string, message string) error {
	return nil
}

func Cache(host string, data any) (string, error) {
	return "", nil
}

func GetFromCache(host string, key string) (any, error) {
	return nil, nil
}

func ProvisionSupervisor(processor string, moduleName, clusterName string, supervisor uint64, config *cluster.Config, metadata map[string]string) error {

	body := &struct {
		Module     string            `json:"module"`
		Cluster    string            `json:"cluster"`
		Config     cluster.Config    `json:"config"`
		Supervisor uint64            `json:"id"`
		Metadata   map[string]string `json:"metadata"`
	}{
		moduleName, clusterName, *config, supervisor, metadata,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	url := fmt.Sprintf("http://%s/supervisor", processor)
	rsp, err := http.Post(url, "application/json", &buf)

	if err != nil {
		return err
	}

	if rsp.Status != "200 OK" {
		return errors.New("failed to provision new supervisor")
	}
	return nil
}

func UpdateSupervisor(host string, id uint64, status supervisor.Status, stats *cluster.Statistics) error {

	url := fmt.Sprintf("%s/supervisor", host)
	client := http.Client{}

	sup := supervisor.Supervisor{
		Id:         id,
		Status:     status,
		Statistics: stats,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(sup)

	req, err := http.NewRequest(http.MethodPut, url, &buf)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.Status != "200 OK" {
		return errors.New("failed to update supervisor")
	}

	return nil
}
