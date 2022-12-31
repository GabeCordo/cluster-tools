package logger

import "time"

type LogOutput int

const (
	Verbose LogOutput = iota
	Simple
)

const (
	dateOutputFormat = "01-02-2006_15:04:05"
)

type Log struct {
	Component string
	Data      string
}

type LogInterval struct {
	Hours        float64   `json:"hours"`
	Minutes      float64   `json:"minutes"`
	LastInterval time.Time `json:"lastInterval"`
}

type ILogger interface {
	Log(template string, params ...interface{})
	Alert(template string, params ...interface{})
	Warning(template string, params ...interface{})
}

type Logger struct {
	Folder   string    `json:"folder"`
	Style    LogOutput `json:"style"`
	LogQueue chan string
	Interval LogInterval `json:"interval"`
}
