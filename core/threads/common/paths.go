package common

import "os"

var (
	userCacheDir, _         = os.UserCacheDir()
	DefaultFrameworkFolder  = userCacheDir + "/etl/"
	DefaultConfigsFolder    = DefaultFrameworkFolder + "configs/"
	DefaultConfigFile       = DefaultFrameworkFolder + "global.etl.yml"
	DefaultLogsFolder       = DefaultFrameworkFolder + "logs/"
	DefaultStatisticsFolder = DefaultFrameworkFolder + "statistics/"
	DefaultSchedulesFolder  = DefaultFrameworkFolder + "schedules/"
)
