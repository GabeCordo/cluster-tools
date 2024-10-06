package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sentmint/cluster-tools/internal/core/database/pipeline"
	"github.com/Sentmint/cluster-tools/internal/core/database/run"
	"net/http"
)

func RunPipelineOnProcessor(host string, pl *pipeline.Pipeline) error {

	body := &struct {
		Namespace  string            `json:"namespace"`
		Supervisor uint64            `json:"id"`
		Config     pipeline.Pipeline `json:"pipeline"`
		Metadata   map[string]string `json:"metadata"`
	}{
		"default", 0, *pl, make(map[string]string),
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/run", host)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code %d", rsp.StatusCode)
	}

	return nil
}

func IsPipelineOnCore(host, namespace, pl string) (bool, error) {

	url := fmt.Sprintf("%s/pipeline?namespace=%s", host, namespace)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	pipelines := make([]pipeline.Pipeline, 0)
	err = json.NewDecoder(rsp.Body).Decode(&pipelines)
	if err != nil {
		return false, err
	}

	for _, p := range pipelines {
		if p.Identifier == pl {
			return true, nil
		}
	}

	return false, nil
}

func CreatePipelineOnCore(host, namespace string, pl *pipeline.Pipeline) error {

	url := fmt.Sprintf("%s/pipeline?namespace=%s", host, namespace)

	b, err := json.Marshal(pl)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	} else {
		return nil
	}
}

func ReplacePipelineOnCore(host, namespace string, pl *pipeline.Pipeline) error {

	url := fmt.Sprintf("%s/pipeline?namespace=%s", host, namespace)

	b, err := json.Marshal(pl)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	} else {
		return nil
	}
}

func RemovePipelineFromCore(host, namespace, pipeline string) error {

	url := fmt.Sprintf("%s/pipeline?namespace=%s&pipeline=%s", host, namespace, pipeline)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	}
	return nil
}

func RunPipelineOnCore(host, namespace, pipeline string) (uint64, error) {

	url := fmt.Sprintf("%s/run", host)

	body := &struct {
		Namespace string            `json:"namespace"`
		Pipeline  string            `json:"pipeline"`
		Metadata  map[string]string `json:"metadata"`
	}{namespace, pipeline, make(map[string]string)}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	response := &struct{ Id uint64 }{}

	if rsp.StatusCode != http.StatusOK {
		return 0, errors.New("http status code " + rsp.Status)
	}

	err = json.NewDecoder(rsp.Body).Decode(response)
	if err != nil {
		return 0, err
	}

	return response.Id, nil
}

func GetRunStatus(host, namespace string, id uint64) (run.Run, error) {

	url := fmt.Sprintf("%s/run?namespace=%s&id=%d", host, namespace, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return run.Run{}, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return run.Run{}, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return run.Run{}, errors.New("http status code " + rsp.Status)
	}

	r := &struct {
		Data []run.Run `json:"data"`
	}{}

	if err = json.NewDecoder(rsp.Body).Decode(r); err != nil {
		return run.Run{}, err
	}

	for _, j := range r.Data {
		if j.Id == id {
			return j, nil
		}
	}

	return run.Run{}, errors.New("run not found")
}
