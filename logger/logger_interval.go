package logger

import (
	"time"
)

func NewInterval(hours, minutes float64) LogInterval {
	now := time.Now()
	return LogInterval{hours, minutes, now}
}

func (li LogInterval) Expired() bool {
	now := time.Now()
	difference := now.Sub(li.lastInterval)
	if difference.Hours() > li.hours {
		if difference.Minutes() > li.minutes {
			return true
		}
	}
	return false
}

func (li LogInterval) Refresh() {
	li.lastInterval = time.Now()
}

func (li LogInterval) String() string {
	return li.lastInterval.Format(dateOutputFormat)
}
