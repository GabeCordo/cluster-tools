package api

import (
	"errors"
	"fmt"
	"github.com/Sentmint/cluster-tools/internal/core/processor"
	"net/http"
)

var PingFailedError = errors.New("ping towards processor failed")

func Probe(processor *processor.Processor) error {

	if processor == nil {
		return errors.New("nil processor")
	}

	url := fmt.Sprintf("http://%s/debug", processor.ToString())
	rsp, err := http.Get(url)
	if err != nil {
		return PingFailedError
	}

	if rsp.StatusCode != http.StatusOK {
		return PingFailedError
	}

	return nil
}
