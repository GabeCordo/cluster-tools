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

func (interval Interval) ToString() string {

	var sb strings.Builder

	sb.WriteString(FormatToCrontab(interval.Minute))
	sb.WriteString(FormatToCrontab(interval.Hour))
	sb.WriteString(FormatToCrontab(interval.Day))
	sb.WriteString(FormatToCrontab(interval.Month))

	return sb.String()
}
