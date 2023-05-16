package cluster

import "time"

func NewResponse(config Config, statistics *Statistics, lapsedTime time.Duration) *Response {
	response := new(Response)

	response.Config = config
	response.Stats = statistics
	response.LapsedTime = lapsedTime

	return response
}
