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
	"time"
)

var client http.Client = http.Client{Timeout: 2 * time.Second}

func ConnectToCore(host string, config *processor.Config) error {

	url := fmt.Sprintf("%s/processor", host)

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(config)

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

	response := &communication.Response{}
	json.NewDecoder(rsp.Body).Decode(response)

	if response.Success == false {
		return errors.New("could not connect to core")
	}

	return err
}

func DisconnectFromCore(host string, config *processor.Config) error {

	url := fmt.Sprintf("%s/processor", host)

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
	defer rsp.Body.Close()

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
