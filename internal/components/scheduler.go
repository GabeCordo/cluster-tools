package components

import "github.com/GabeCordo/cluster-tools/internal/interfaces"

type Scheduler interface {
	GetQueue() []interfaces.Job
	ItemsInQueue() int
	Print()
}
