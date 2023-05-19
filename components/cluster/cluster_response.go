package cluster

import "time"

func NewResponse(config Config, statistics *Statistics, lapsedTime time.Duration, crashed bool) *Response {
	response := new(Response)

	response.Config = config
	response.Stats = statistics
	response.LapsedTime = lapsedTime
	response.DidItCrash = crashed

	return response
}
