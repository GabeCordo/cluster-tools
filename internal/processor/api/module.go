package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/clarence/internal/interfaces"
	"net/http"
	"strconv"
)

func CreateModule(host string, processor *interfaces.ProcessorConfig, config *interfaces.ModuleConfig) error {

	url := fmt.Sprintf("%s/module", host)

	request := &interfaces.HTTPRequest{
		Host: processor.Host,
		Port: processor.Port,
		Module: interfaces.HTTPModuleRequest{
			Config: *config,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}

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
		return errors.New("unexpected response code")
	}

	response := &interfaces.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	if !response.Success {
		return errors.New(response.Description)
	}

	return nil
}

func DeleteModule(host string, processor *interfaces.ProcessorConfig, module *interfaces.ModuleConfig) error {
	url := fmt.Sprintf("%s/module", host)

	req, err := http.NewRequest(http.MethodDelete, url, nil)

	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("host", processor.Host)
	q.Add("port", strconv.Itoa(processor.Port))
	q.Add("module", module.Name)

	req.URL.RawQuery = q.Encode()

	rsp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.Status != "200 OK" {
		return errors.New("unexpected response code")
	}

	response := &interfaces.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	if response.Success == false {
		return errors.New("could not disconnect from core")
	}

	return err
}
