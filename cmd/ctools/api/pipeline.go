package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"net/http"
)

func RunPipelineOnProcessor(pl *pipeline.Pipeline) error {

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

	url := "http://localhost:5023/run"
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

func IsPipelineOnCore(namespace, pipeline string) {

}

func CreatePipelineToCore(namespace, pl *pipeline.Pipeline) {

}

func ReplacePipelineOnCore(namespace, pl *pipeline.Pipeline) {

}

func RunPipelineOnCore() {

}
