package text_logging

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/Flock/internal/shared/logging"
	"github.com/GabeCordo/Flock/internal/shared/terminal"
	"log"
)

//////////////////////////////////////////////////////////////////////////////
//							    Text Logging
//////////////////////////////////////////////////////////////////////////////

/* ------------------------ **** Constants ***** -------------------------- */

const LoggerFormat = "[%s][%s] "

/* ----------------------- **** Static Vars ***** ------------------------- */

var DefaultDebugValue = true

/* ------------------------ ****   Types   ***** -------------------------- */

type TextLogger struct {
	thread    string
	colour    string
	colourItr int
	debug     *bool
}

/* ------------------------ **** Functions ***** -------------------------- */

func New(thread string, debug ...*bool) (*TextLogger, error) {

	logger := new(TextLogger)

	logger.thread = thread
	logger.colour = terminal.White

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

func (logger *TextLogger) prefix(alert logging.AlertType) string {
	format := fmt.Sprintf(LoggerFormat, logger.thread, alert.ToString())
	return logger.colour + format + terminal.Reset
}

func (logger *TextLogger) canPrint() bool {
	return *(logger.debug) == true
}

func (logger *TextLogger) SwapColour() {
	if logger.colourItr == (terminal.NumOfColours - 1) {
		logger.colourItr = 0
	} else {
		logger.colourItr++
	}

	logger.colour = terminal.Colours[logger.colourItr]
}

func (logger *TextLogger) SetColour(colour string) {
	logger.colour = colour
}

func (logger *TextLogger) Println(text string) {
	if !logger.canPrint() {
		return
	}
	log.Println(logger.prefix(logging.Normal) + text)
}

func (logger *TextLogger) Print(text string) {
	if !logger.canPrint() {
		return
	}
	log.Print(logger.prefix(logging.Normal) + text)
}

func (logger *TextLogger) Printf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Printf(logger.prefix(logging.Normal)+template, other...)
}

func (logger *TextLogger) Warn(text string) {
	if !logger.canPrint() {
		return
	}
	log.Print(logger.prefix(logging.Warning) + text)
}

func (logger *TextLogger) Warnln(text string) {
	if !logger.canPrint() {
		return
	}
	log.Println(logger.prefix(logging.Warning) + text)
}

func (logger *TextLogger) Warnf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Printf(logger.prefix(logging.Warning)+template, other...)
}

func (logger *TextLogger) Alert(text string) {
	if !logger.canPrint() {
		return
	}
	log.Print(logger.prefix(logging.Alert) + text)
}

func (logger *TextLogger) Alertln(text string) {
	if !logger.canPrint() {
		return
	}
	log.Println(logger.prefix(logging.Alert) + text)
}

func (logger *TextLogger) Alertf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Printf(logger.prefix(logging.Alert)+template, other...)
}

func (logger *TextLogger) Panic(text string) {
	if !logger.canPrint() {
		return
	}
	log.Panic(logger.prefix(logging.Alert) + text)
}

func (logger *TextLogger) Panicln(text string) {
	if !logger.canPrint() {
		return
	}
	log.Panicln(logger.prefix(logging.Panic) + text)
}

func (logger *TextLogger) Panicf(template string, other ...any) {
	if !logger.canPrint() {
		return
	}
	log.Panicf(logger.prefix(logging.Panic)+template, other...)
}
