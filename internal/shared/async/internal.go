package async

type Action uint8

const (
	Create Action = iota
	Update
	Delete
)

type Record uint8

const (
	Module Record = iota
	Log
	Run
)

type Request struct {
	Action Action `json:"action"`
	Record Record `json:"record"`
	Data   any    `json:"data"`
}
