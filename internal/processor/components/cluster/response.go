package cluster

import (
	"github.com/GabeCordo/clarence/cluster"
	"github.com/GabeCordo/clarence/internal/interfaces"
	"time"
)

func NewResponse(config cluster.Config, statistics *interfaces.Statistics, lapsedTime time.Duration, crashed bool) *Response {
	response := new(Response)

	response.Config = config
	response.Stats = statistics
	response.LapsedTime = lapsedTime
	response.DidItCrash = crashed

	return response
}
