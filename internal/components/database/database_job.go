package database

import "github.com/GabeCordo/cluster-tools/internal/core/interfaces"

type JobDatabase interface {
	GetAll() ([]interfaces.Job, error)
	GetBy(filter *interfaces.Filter) ([]interfaces.Job, error)
	Create(job *interfaces.Job) error
	Delete(filter *interfaces.Filter) error
}
