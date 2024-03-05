package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/core/components/messenger"
	"github.com/GabeCordo/cluster-tools/core/components/supervisor"
	"github.com/GabeCordo/cluster-tools/core/interfaces/cluster"
	"github.com/GabeCordo/cluster-tools/core/interfaces/communication"
	"net/http"
)

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

func UpdateSupervisor(host string, id uint64, status supervisor.Status, stats *cluster.Statistics) error {

	url := fmt.Sprintf("%s/supervisor", host)

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
	defer rsp.Body.Close()

	if rsp.Status != "200 OK" {
		return errors.New("failed to update supervisor")
	}

	return nil
}

func Cache(host string, key string, data string) (string, error) {

	url := fmt.Sprintf("%s/cache", host)

	body := &struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key: key, Value: data,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		// TODO : remove static value
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.Status != "200 OK" {
		return "", errors.New("could not store in cache")
	}

	response := &communication.Response{}
	err = json.NewDecoder(rsp.Body).Decode(response)
	if err != nil {
		return "", err
	}

	return (response.Data).(string), err
}

func GetFromCache(host string, key string) (string, error) {

	url := fmt.Sprintf("%s/cache", host)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("key", key)
	req.URL.RawQuery = q.Encode()

	rsp, err := client.Do(req)

	if err != nil {
		return "", errors.New("could not reach cache")
	}
	defer rsp.Body.Close()

	if rsp.Status != "200 OK" {
		return "", errors.New("cache not found")
	}

	response := &communication.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	return (response.Data).(string), nil
}

func log(host string, id uint64, level messenger.MessagePriority, message string) error {

	url := fmt.Sprintf("%s/log", host)

	data := &supervisor.Log{Id: id, Level: level, Message: message}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)

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
		return errors.New("was not able to send a log")
	}

	return nil
}

func Log(host string, id uint64, message string) error {

	//return log(host, id, messenger.Normal, message)
	return nil
}

func LogWarn(host string, id uint64, message string) error {

	//return log(host, id, messenger.Warning, message)
	return nil
}

func LogError(host string, id uint64, message string) error {

	//return log(host, id, messenger.Fatal, message)
	return nil
}
