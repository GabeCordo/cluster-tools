package pipeline

type Action uint8

const (
	Next Action = iota
	Span
	Join
)

type Pipeline struct {
}

type Stage struct {
}
