package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FortifiedCode/plover"
	"io"
	"net/http"

	"github.com/FortifiedCode/flock/internal/core/database/run"
)

func RunPipelineOnProcessor(host string, pl *plover.PipelineIR) error {

	body := &struct {
		Namespace  string            `json:"namespace"`
		Supervisor uint64            `json:"id"`
		Config     plover.PipelineIR `json:"pipeline"`
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	pipelines := make([]plover.PipelineIR, 0)
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

func CreatePipelineOnCore(host, namespace string, pl *plover.PipelineIR) error {

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	} else {
		return nil
	}
}

func ReplacePipelineOnCore(host, namespace string, pl *plover.PipelineIR) error {

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

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

func GetRunStatus(host, namespace string, id uint64) (*run.Run, error) {

	url := fmt.Sprintf("%s/run?namespace=%s&id=%d", host, namespace, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return nil, errors.New("http status code " + rsp.Status)
	}

	r := &struct {
		Data []*run.Run `json:"data"`
	}{}

	if err = json.NewDecoder(rsp.Body).Decode(r); err != nil {
		return nil, err
	}

	for _, j := range r.Data {
		if j.Id == id {
			return j, nil
		}
	}

	return nil, errors.New("run not found")
}

func StopRun(host string, id uint64) error {

	url := fmt.Sprintf("%s/run?id=%d", host, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return errors.New("something went wrong while stopping the run")
	} else {
		return nil
	}
}
