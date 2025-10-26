package pipeline

import "github.com/FortifiedCode/plover"

type Pipeline struct {
	Namespace  string             `json:"namespace"`
	Identifier string             `json:"identifier"`
	Data       *plover.PipelineIR `json:"data"`
}
