package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/FortifiedCode/flock/internal/targets/core/component/processor"
)

type ProcessorsResponse struct {
	Success     bool                  `json:"success"`
	Description string                `json:"description"`
	Processors  []processor.Processor `json:"data"`
}

func GetProcessors(host string) ([]processor.Processor, error) {

	url := fmt.Sprintf("%s/processor", host)

	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	response := ProcessorsResponse{}
	err = json.NewDecoder(rsp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		err = errors.New(response.Description)
	}

	return response.Processors, err
}
