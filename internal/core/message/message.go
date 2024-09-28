package message

type Priority string

const (
	Normal  Priority = "normal"
	Warning          = "warning"
	Fatal            = "fatal"
	Any              = ""
)

func (priority Priority) Shortform() string {
	if priority == Normal {
		return "-"
	} else if priority == Warning {
		return "?"
	} else if priority == Fatal {
		return "!"
	} else {
		return ""
	}
}

func FromShortform(shortform string) Priority {
	if shortform == "-" {
		return Normal
	} else if shortform == "?" {
		return Warning
	} else if shortform == "!" {
		return Fatal
	} else {
		return Any
	}
}

type Source struct {
	Module     string
	Cluster    string
	Identifier uint64
}

type Messenger interface {
	Message(source Source, record any) error
	Flush(source Source, destination any) error
}
