package core

import "os"

const MongoDatabaseUriEnv = "MONGO_DATABASE_URI"

type EnvironmentVariables struct {
	MongoDbUri string
}

func ReadEnvironmentVariables() (env EnvironmentVariables) {

	env.MongoDbUri = os.Getenv(MongoDatabaseUriEnv)

	return env
}
