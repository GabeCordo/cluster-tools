package core

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/components/utils"
)

type Core struct {
	HttpThread        *HttpThread
	ProvisionerThread *ProvisionerThread
	MessengerThread   *MessengerThread
	DatabaseThread    *DatabaseThread
	CacheThread       *CacheThread

	C1        chan threads.DatabaseRequest
	C2        chan threads.DatabaseResponse
	C3        chan threads.MessengerRequest
	C4        chan threads.MessengerResponse
	C5        chan threads.ProvisionerRequest
	C6        chan threads.ProvisionerResponse
	C7        chan threads.DatabaseRequest
	C8        chan threads.DatabaseResponse
	C9        chan threads.CacheRequest
	C10       chan threads.CacheResponse
	C11       chan threads.MessengerRequest
	interrupt chan threads.InterruptEvent

	logger *utils.Logger
}
