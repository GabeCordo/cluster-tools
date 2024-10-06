package scheduler

import (
	"github.com/Sentmint/cluster-tools/internal/core/database/job"
)

type Schedule struct {
	Minute int `json:"minute", yaml:"minute"`
	Hour   int `json:"hour", yaml:"hour"`
	Day    int `json:"day", yaml:"day"`
	Month  int `json:"month", yaml:"month"`
}

type Job struct {
	Cluster  string   `json:"cluster", yaml:"cluster"`
	Config   string   `json:"pipeline", yaml:"pipeline"`
	Schedule Schedule `json:"scheduler", yaml:"scheduler"`
}

type Scheduler interface {
	GetQueue() []job.Job
	ItemsInQueue() int
	Print()
}
