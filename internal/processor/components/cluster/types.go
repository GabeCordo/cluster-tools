package cluster

import (
	"github.com/GabeCordo/clarence/cluster"
	"github.com/GabeCordo/clarence/internal/interfaces"
	"time"
)

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

type Transformer func(in []any) (out any, success bool)

type Lambda func(helper cluster.H, metadata cluster.M) (out any, success bool)

type LoadAll interface {
	LoadFunc(helper cluster.H, metadata cluster.M, in []any)
}

type LoadOne interface {
	LoadFunc(helper cluster.H, metadata cluster.M, in any)
}

type SystemFunctions interface {
	Setup(curr time.Time, h cluster.H)
	Teardown(curr time.Time, h cluster.H)
}

type VerifiableET interface {
	VerifyETFunction(in any) (valid bool)
}

type VerifiableTL interface {
	VerifyTLFunction(in any) (valid bool)
}

// Test
// TODO : needs to be implemented
type Test interface {
	MockExtractFunc(metadata cluster.M, out cluster.Out)
	VerifyTransformOutput(metadata cluster.M, in any) (success bool)
	MockLoadFunc(metadata cluster.M, in any)
}

type Event uint8

const (
	Register = iota
	Mount
	UnMount
	Use
	Delete
)

type Response struct {
	Config     cluster.Config         `json:"core"`
	Stats      *interfaces.Statistics `json:"stats"`
	LapsedTime time.Duration          `json:"lapsed-time"`
	DidItCrash bool                   `json:"crashed"`
}
