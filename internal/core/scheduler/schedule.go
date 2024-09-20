package scheduler

import (
	"github.com/GabeCordo/cluster-tools/internal/database/job"
)

type Schedule struct {
	Minute int `json:"minute", yaml:"minute"`
	Hour   int `json:"hour", yaml:"hour"`
	Day    int `json:"day", yaml:"day"`
	Month  int `json:"month", yaml:"month"`
}

type Job struct {
	Cluster  string   `json:"cluster", yaml:"cluster"`
	Config   string   `json:"config", yaml:"config"`
	Schedule Schedule `json:"scheduler", yaml:"scheduler"`
}

type Scheduler interface {
	GetQueue() []job.Job
	ItemsInQueue() int
	Print()
}
