package database

import "github.com/GabeCordo/cluster-tools/core/components/database"

var statisticDatabase *database.StatisticDatabase

func GetStatisticDatabaseInstance() *database.StatisticDatabase {
	if statisticDatabase == nil {
		statisticDatabase = database.NewStatisticDatabase()
	}
	return statisticDatabase
}

var configDatabase *database.ConfigDatabase

func GetConfigDatabaseInstance() *database.ConfigDatabase {
	if configDatabase == nil {
		configDatabase = database.NewConfigDatabase()
	}
	return configDatabase
}
