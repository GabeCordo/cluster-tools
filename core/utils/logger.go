package utils

import (
	"errors"
	"fmt"
	"log"
)

type ThreadType uint8

const (
	HttpClient ThreadType = iota
	HttpProcessor
	Provisioner
	Processor
	Supervisor
	Database
	Messenger
	Cache
	Undefined
)

func (threadType ThreadType) ToString() string {
	switch threadType {
	case HttpClient:
		return "HTTP"
	case Provisioner:
		return "PROVISIONER"
	case Messenger:
		return "MESSENGER"
	case Database:
		return "DATABASE"
	case Cache:
		return "CACHE"
	default:
		return "-"
	}
}

type AlertType uint8

const (
	Normal AlertType = iota
	Warning
	Alert
	Panic
)

func (alertType AlertType) ToString() string {
	switch alertType {
	case Warning:
		return "?"
	case Alert:
		return "!"
	case Panic:
		return "p"
	default:
		return "-"
	}
}

var (
	DefaultDebugValue = true
	LoggerFormat      = "[%s][%s] "
)

type Logger struct {
	thread    ThreadType
	colour    string
	colourItr int

	debug *bool
}

func NewLogger(thread ThreadType, debug ...*bool) (*Logger, error) {

	logger := new(Logger)

	logger.thread = thread
	logger.colour = White

	if len(debug) < 1 {
		logger.debug = &DefaultDebugValue
	} else {
		if debug[0] == nil {
			return nil, errors.New("the debug bool pointer was nil")
		}
		logger.debug = debug[0]
	}

	return logger, nil
}

func (logger *Logger) prefix(alert AlertType) string {
	format := fmt.Sprintf(LoggerFormat, logger.thread.ToString(), alert.ToString())
	return logger.colour + format + Reset
}

func (logger *Logger) canPrint() bool {
	return *(logger.debug) == true
}

func (logger *Logger) SwapColour() {
	if logger.colourItr == (NumOfColours - 1) {
		logger.colourItr = 0
	} else {
		logger.colourItr++
	}

	logger.colour = Colours[logger.colourItr]
}

func (logger *Logger) SetColour(colour string) {
	logger.colour = colour
}

func (logger *Logger) Println(text string) {
	if !logger.canPrint() {
		return
	}
	log.Println(logger.prefix(Normal) + text)
}

func (logger *Logger) Print(text string) {
	if !logger.canPrint() {
		return
	}
	log.Print(logger.prefix(Normal) + text)
}

func (logger *Logger) Printf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Printf(logger.prefix(Normal)+template, other...)
}

func (logger *Logger) Warn(text string) {
	if !logger.canPrint() {
		return
	}
	log.Print(logger.prefix(Warning) + text)
}

func (logger *Logger) Warnln(text string) {
	if !logger.canPrint() {
		return
	}
	log.Println(logger.prefix(Warning) + text)
}

func (logger *Logger) Warnf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Printf(logger.prefix(Warning)+template, other...)
}

func (logger *Logger) Alert(text string) {
	if !logger.canPrint() {
		return
	}
	log.Print(logger.prefix(Alert) + text)
}

func (logger *Logger) Alertln(text string) {
	if !logger.canPrint() {
		return
	}
	log.Println(logger.prefix(Alert) + text)
}

func (logger *Logger) Alertf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Printf(logger.prefix(Alert)+template, other...)
}

func (logger *Logger) Panic(text string) {
	if !logger.canPrint() {
		return
	}
	log.Panic(logger.prefix(Alert) + text)
}

func (logger *Logger) Panicln(text string) {
	if !logger.canPrint() {
		return
	}
	log.Panicln(logger.prefix(Panic) + text)
}

func (logger *Logger) Panicf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Panicf(logger.prefix(Panic)+template, other...)
}
