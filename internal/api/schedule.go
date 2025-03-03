package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sentmint/pops/internal/core/database/job"
	"net/http"
)

func GetJobs(host, namespace string) ([]job.Job, error) {

	url := fmt.Sprintf("%s/job?namespace=%s", host, namespace)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := struct{ Data []job.Job }{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func CreateJob(host string, job job.Job) error {

	url := fmt.Sprintf("%s/job", host)
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	} else {
		return nil
	}
}

func DeleteJob(host, job string) error {

	url := fmt.Sprintf("%s/job?id=%s", host, job)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	} else {
		return nil
	}
}
