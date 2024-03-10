package common

import "os"

var (
	userCacheDir, _         = os.UserCacheDir()
	DefaultFrameworkFolder  = userCacheDir + "/cluster.tools/"
	DefaultConfigsFolder    = DefaultFrameworkFolder + "configs/"
	DefaultConfigFile       = DefaultFrameworkFolder + "global.ct.yml"
	DefaultLogsFolder       = DefaultFrameworkFolder + "logs/"
	DefaultStatisticsFolder = DefaultFrameworkFolder + "statistics/"
	DefaultSchedulesFolder  = DefaultFrameworkFolder + "schedules/"
)
