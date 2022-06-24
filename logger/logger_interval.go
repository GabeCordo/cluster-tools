package logger

import (
	"time"
)

func NewInterval(hours, minutes float64) LogInterval {
	now := time.Now()
	return LogInterval{hours, minutes, now}
}

func (li *LogInterval) Expired() bool {
	now := time.Now()
	difference := now.Sub(li.LastInterval)
	if difference.Hours() > li.Hours {
		if difference.Minutes() > li.Minutes {
			return true
		}
	}
	return false
}

func (li *LogInterval) Refresh() {
	li.LastInterval = time.Now()
}

func (li *LogInterval) String() string {
	return li.LastInterval.Format(dateOutputFormat)
}
