package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Gateway() (string, error) {

	rsp, err := http.Get("http://localhost:5023/core")
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", nil
	}

	r := Response{}
	err = json.NewDecoder(rsp.Body).Decode(&r)
	if err != nil {
		return "", err
	}

	gateway, ok := r.Data.(string)
	if !ok {
		return "", fmt.Errorf("gateway response is not a string")
	}

	return gateway, nil
}

func Connect(core string) error {

	url := fmt.Sprintf("http://localhost:5023/gate?=%saction&=connect", core)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return nil
}

func Disconnect(core string) error {

	url := fmt.Sprintf("http://localhost:5023/gate?=%saction&=disconnect", core)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return nil
}
