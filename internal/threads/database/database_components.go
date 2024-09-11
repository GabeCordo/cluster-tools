package database

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/database"
	"os"
)

var mongoUrl = os.Getenv("MONGO_DB_URL")

var statisticDatabase database.StatisticDatabase

func GetStatisticDatabaseInstance(t ...string) database.StatisticDatabase {
	if statisticDatabase == nil {

		if len(t) == 0 || (len(t) == 1 && t[0] == "file") {
			statisticDatabase = database.NewLocalStatisticDatabase()
		} else {
			statisticDatabase, _ = database.NewMongoStatisticsDatabase(mongoUrl)
		}
	}
	return statisticDatabase
}

var configDatabase database.ConfigDatabase

func GetConfigDatabaseInstance(t ...string) database.ConfigDatabase {
	if configDatabase == nil {

		if len(t) == 0 || (len(t) == 1 && t[0] == "file") {
			configDatabase = database.NewLocalConfigDatabase()
		} else {
			configDatabase, _ = database.NewMongoConfigDatabase(mongoUrl)
		}
	}
	return configDatabase
}

var jobDatabase database.JobDatabase

func GetJobDatabaseInstance(t ...string) database.JobDatabase {
	if jobDatabase == nil {

		if len(t) == 0 || (len(t) == 1 && t[0] == "file") {
			jobDatabase = database.NewLocalJobDatabase()
		} else {
			jobDatabase, _ = database.NewMongoJobDatabase(mongoUrl)
		}
	}
	return jobDatabase
}
