package threads

import (
	"github.com/GabeCordo/etl-light/threads"
	"github.com/GabeCordo/etl/core/threads/cache"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/threads/database"
	"github.com/GabeCordo/etl/core/threads/http_client"
	"github.com/GabeCordo/etl/core/threads/http_processor"
	"github.com/GabeCordo/etl/core/threads/messenger"
	"github.com/GabeCordo/etl/core/threads/processor"
	"github.com/GabeCordo/etl/core/threads/supervisor"
	"github.com/GabeCordo/etl/core/utils"
	"os"
)

var (
	userCacheDir, _         = os.UserCacheDir()
	DefaultFrameworkFolder  = userCacheDir + "/etl/"
	DefaultModulesFolder    = DefaultFrameworkFolder + "modules/"
	DefaultConfigsFolder    = DefaultFrameworkFolder + "configs/"
	DefaultConfigFile       = DefaultFrameworkFolder + "global.etl.yml"
	DefaultLogsFolder       = DefaultFrameworkFolder + "logs/"
	DefaultStatisticsFolder = DefaultFrameworkFolder + "statistics/"
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
	C7        chan threads.DatabaseRequest
	C8        chan threads.DatabaseResponse
	C9        chan threads.CacheRequest
	C10       chan threads.CacheResponse
	C11       chan threads.MessengerRequest
	C12       chan common.ProcessorRequest
	C13       chan common.ProcessorResponse
	C14       chan common.SupervisorRequest
	C15       chan common.SupervisorResponse
	interrupt chan threads.InterruptEvent

	logger *utils.Logger
}
