package logging

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
		return "?"
	case Alert:
		return "!"
	case Panic:
		return "p"
	default:
		return "-"
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
