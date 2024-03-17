package scheduler

import (
	"fmt"
	"strings"
)

const (
	DefaultPlaceholder = "*"
)

func FormatToCrontab(value int) string {

	postfix := ""
	if (value != 0) && (value != 60) {
		postfix = fmt.Sprintf("/%d", value)
	}

	return DefaultPlaceholder + postfix + " "
}

func (interval Interval) Empty() bool {

	return (interval.Month == 0) && (interval.Day == 0) && (interval.Hour == 0) && (interval.Minute == 0)
}

func (interval Interval) Equals(other *Interval) bool {

	if other == nil {
		return false
	}

	return interval.Hour == other.Hour &&
		interval.Day == other.Day &&
		interval.Month == other.Month &&
		interval.Minute == other.Minute
}

func (interval Interval) ToString() string {

	var sb strings.Builder

	sb.WriteString(FormatToCrontab(interval.Minute))
	sb.WriteString(FormatToCrontab(interval.Hour))
	sb.WriteString(FormatToCrontab(interval.Day))
	sb.WriteString(FormatToCrontab(interval.Month))

	return sb.String()
}
