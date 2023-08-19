package threads

import (
	"github.com/GabeCordo/mango-core/core/threads/cache"
	"github.com/GabeCordo/mango-core/core/threads/common"
	"github.com/GabeCordo/mango-core/core/threads/database"
	"github.com/GabeCordo/mango-core/core/threads/http_client"
	"github.com/GabeCordo/mango-core/core/threads/http_processor"
	"github.com/GabeCordo/mango-core/core/threads/messenger"
	"github.com/GabeCordo/mango-core/core/threads/processor"
	"github.com/GabeCordo/mango-core/core/threads/supervisor"
	"github.com/GabeCordo/mango/threads"
	"github.com/GabeCordo/mango/utils"
)

type Core struct {
	HttpClientThread    *http_client.Thread
	HttpProcessorThread *http_processor.Thread
	ProcessorThread     *processor.Thread
	SupervisorThread    *supervisor.Thread
	MessengerThread     *messenger.Thread
	DatabaseThread      *database.Thread
	CacheThread         *cache.Thread

	C1        chan threads.DatabaseRequest
	C2        chan threads.DatabaseResponse
	C3        chan threads.MessengerRequest
	C4        chan threads.MessengerResponse
	C5        chan common.ProcessorRequest
	C6        chan common.ProcessorResponse
	C7        chan common.ProcessorRequest
	C8        chan common.ProcessorResponse
	C9        chan threads.CacheRequest
	C10       chan threads.CacheResponse
	C11       chan threads.DatabaseRequest
	C12       chan threads.DatabaseResponse
	C13       chan common.SupervisorRequest
	C14       chan common.SupervisorResponse
	C15       chan threads.DatabaseRequest
	C16       chan threads.DatabaseResponse
	C17       chan threads.MessengerRequest
	interrupt chan threads.InterruptEvent

	logger *utils.Logger
}