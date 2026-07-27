package pipeline

import (
	"github.com/GabeCordo/FunctionScheduler/internal/targets/core/database"
)

type Pipeline struct {
	Namespace  string                       `json:"namespace"`
	Identifier string                       `json:"identifier"`
	Data       *ScalingFunctions.PipelineIR `json:"data"`
}

type Database interface {
	Get(filter database.Filter) []*ScalingFunctions.PipelineIR
	Create(filter database.Filter, record *ScalingFunctions.PipelineIR) (string, error)
	Replace(filter database.Filter, record *ScalingFunctions.PipelineIR) error
	Delete(filter database.Filter) error
	Distinct(filter database.Filter) ([]any, error)
	Save(path string) error
	Load(path string) error
	Print()
}
