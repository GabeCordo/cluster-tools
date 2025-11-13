package logging

import "github.com/FortifiedCode/flock/internal/shared/terminal"

//////////////////////////////////////////////////////////////////////////////
//							   	  Logging
//////////////////////////////////////////////////////////////////////////////

/* ------------------------ ****   Enums   ***** -------------------------- */

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
		return terminal.Yellow + "@" + terminal.Reset
	case Alert:
		return terminal.Orange + "!" + terminal.Reset
	case Panic:
		return terminal.Red + "p" + terminal.Reset
	default:
		return terminal.White + "+" + terminal.Reset
	}
}

/* ------------------------ ****   Types   ***** -------------------------- */

type Logger interface {
	SetColour(colour string)
	Println(text string)
	Print(text string)
	Printf(template string, other ...any)
	Warn(text string)
	Warnln(text string)
	Warnf(template string, other ...any)
	Alert(text string)
	Alertln(text string)
	Alertf(template string, other ...any)
	Panic(text string)
	Panicln(text string)
	Panicf(template string, other ...any)
}
