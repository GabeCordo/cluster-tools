package scheduler

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/database"
	"os"
)

var mongoUrl = os.Getenv("MONGO_DB_URL")

var jobDatabase database.JobDatabase

func GetJobDatabaseInstance() database.JobDatabase {
	if jobDatabase == nil {
		jobDatabase, _ = database.NewMongoJobDatabase(mongoUrl)
	}
	return jobDatabase
}
