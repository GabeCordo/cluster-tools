package threads

import (
	"github.com/GabeCordo/etl-light/core/threads"
	"github.com/GabeCordo/etl/core/threads/cache"
	"github.com/GabeCordo/etl/core/threads/database"
	"github.com/GabeCordo/etl/core/threads/http"
	"github.com/GabeCordo/etl/core/threads/messenger"
	"github.com/GabeCordo/etl/core/threads/provisioner"
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
	HttpThread        *http.Thread
	ProvisionerThread *provisioner.Thread
	MessengerThread   *messenger.Thread
	DatabaseThread    *database.Thread
	CacheThread       *cache.Thread

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
