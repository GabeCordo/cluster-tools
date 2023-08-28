package api

import "github.com/GabeCordo/mango/core/interfaces/cluster"

func Log(host string, message string) error {
	return nil
}

func LogWarn(host string, message string) error {
	return nil
}

func LogError(host string, message string) error {
	return nil
}

func Cache(host string, data any) (string, error) {
	return "", nil
}

func GetFromCache(host string, key string) (any, error) {
	return nil, nil
}

func SupervisorDone(host string, supervisor uint64, statistics *cluster.Statistics) error {
	return nil
}
