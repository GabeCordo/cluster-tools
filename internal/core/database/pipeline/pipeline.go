package pipeline

type Segment int8

const (
	Extract   Segment = 0
	Transform         = 1
	Load              = 2
)

type OnCrash string

const (
	Restart   OnCrash = "Restart"
	DoNothing         = "DoNothing"
)

type OnLoad string

const (
	CompleteAndPush OnLoad = "CompleteAndPush"
	WaitAndPush            = "WaitAndPush"
)

type RunMode string

const (
	Batch  RunMode = "Batch"
	Stream         = "Stream"
)

type Function struct {
	Module     string `json:"module"`
	Identifier string `json:"id"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	StartWith  int    `json:"start_with"`
}

type Pipe struct {
	Identifier   string  `json:"id"`
	Threshold    int     `json:"threshold"`
	GrowthFactor float64 `json:"growth_factor"`
}

type Pipeline struct {
	Identifier string     `json:"id"`
	OnCrash    OnCrash    `json:"on-crash"`
	Functions  []Function `json:"functions"`
	Pipes      []Pipe     `json:"pipes"`
}
