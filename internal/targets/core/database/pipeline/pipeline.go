package pipeline

import (
	"github.com/FortifiedCode/flock/internal/targets/core/database"
	"github.com/FortifiedCode/plover"
)

type Pipeline struct {
	Namespace  string             `json:"namespace"`
	Identifier string             `json:"identifier"`
	Data       *plover.PipelineIR `json:"data"`
}

type Database interface {
	Get(filter database.Filter) []*plover.PipelineIR
	Create(filter database.Filter, record *plover.PipelineIR) (string, error)
	Replace(filter database.Filter, record *plover.PipelineIR) error
	Delete(filter database.Filter) error
	Distinct(filter database.Filter) ([]any, error)
	Save(path string) error
	Load(path string) error
	Print()
}
