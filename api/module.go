package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/mango/core/interfaces/communication"
	"github.com/GabeCordo/mango/core/interfaces/module"
	processor_i "github.com/GabeCordo/mango/core/interfaces/processor"
	"net/http"
	"strconv"
)

func CreateModule(host string, processor *processor_i.Config, config *module.Config) error {

	url := fmt.Sprintf("%s/module", host)

	request := &communication.HTTPRequest{
		Host: processor.Host,
		Port: processor.Port,
		Module: communication.HTTPModuleRequest{
			Config: *config,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}

	rsp, err := http.Post(url, "application/json", &buf)

	if rsp.Status != "200 OK" {
		return errors.New("unexpected response code")
	}

	response := &communication.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	if !response.Success {
		return errors.New(response.Description)
	}

	return nil
}

func DeleteModule(host string, processor *processor_i.Config, module *module.Config) error {
	url := fmt.Sprintf("%s/module", host)

	client := &http.Client{}

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

	if rsp.Status != "200 OK" {
		return errors.New("unexpected response code")
	}

	response := &communication.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	if response.Success == false {
		return errors.New("could not disconnect from core")
	}

	return err
}
