package channel

import "time"

type DataTimer struct {
	In  time.Time
	Out time.Time
}

type OneWay interface {
	Push(data any)
}
