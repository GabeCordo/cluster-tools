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
	component string
	data      string
}

type LogInterval struct {
	hours        float64
	minutes      float64
	lastInterval time.Time
}

type ILogger interface {
	Log(template string, params ...interface{})
	Alert(template string, params ...interface{})
	Warning(template string, params ...interface{})
}

type Logger struct {
	folder   string
	style    LogOutput
	logQueue chan string
	interval *LogInterval
}
