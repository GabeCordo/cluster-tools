package pipeline

import "github.com/FortifiedCode/plover"

type Wrapper struct {
	Namespace  string            `json:"namespace"`
	Identifier string            `json:"identifier"`
	Pipeline   plover.PipelineIR `json:"pipeline"`
}
