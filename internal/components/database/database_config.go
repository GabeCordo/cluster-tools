package database

import "github.com/GabeCordo/cluster-tools/internal/core/interfaces"

type ConfigDatabase interface {
	Get(filter ConfigFilter) (records []interfaces.Config, err error)
	Create(moduleIdentifier, configIdentifier string, cfg interfaces.Config) (err error)
	Replace(moduleIdentifier, configIdentifier string, cfg interfaces.Config) (err error)
	Delete(moduleIdentifier, configIdentifier string) (err error)
}
