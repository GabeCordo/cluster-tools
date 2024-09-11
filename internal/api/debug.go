package api

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/core/components/processor"
	"net/http"
)

var PingFailedError = errors.New("ping towards processor failed")

func Probe(processor *processor.Processor) error {

	if processor == nil {
		return errors.New("nil processor")
	}

	url := fmt.Sprintf("http://%s:%d/debug", processor.Host, processor.Port)
	rsp, err := http.Get(url)
	if err != nil {
		return PingFailedError
	}

	if rsp.StatusCode != http.StatusOK {
		return PingFailedError
	}

	return nil
}
