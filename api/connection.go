package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GabeCordo/mango/core/interfaces/communication"
	"github.com/GabeCordo/mango/core/interfaces/processor"
	"net/http"
	"strconv"
)

func ConnectToCore(host string, config *processor.Config) error {

	url := fmt.Sprintf("%s/processor", host)

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(config)

	rsp, err := http.Post(url, "application/json", &buf)

	if err != nil {
		return err
	}

	if rsp.Status != "200 OK" {
		return errors.New("unexpected response code")
	}

	response := &communication.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	if response.Success == false {
		return errors.New("could not connect to core")
	}

	return err
}

func DisconnectFromCore(host string, config *processor.Config) error {

	url := fmt.Sprintf("%s/processor", host)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, url, nil)

	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("host", config.Host)
	q.Add("port", strconv.Itoa(config.Port))

	req.URL.RawQuery = q.Encode()

	rsp, err := client.Do(req)

	if err != nil {
		return err
	}

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

func HeartbeatToCore(host string) error {
	return nil
}
