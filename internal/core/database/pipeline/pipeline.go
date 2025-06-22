package pipeline

type Function struct {
	Module     string `json:"module"`
	Identifier string `json:"id" yaml:"id"`
	From       string `json:"from,omitempty" yaml:"from,omitempty"`
	To         string `json:"to,omitempty" yaml:"to,omitempty"`
	StartWith  int    `json:"start_with,omitempty" yaml:"start_with,omitempty"`
	WaitBefore bool   `json:"wait_before,omitempty" yaml:"wait_before"`
}

type Pipe struct {
	Identifier   string  `json:"id" yaml:"id"`
	Threshold    int     `json:"threshold" yaml:"threshold"`
	GrowthFactor float64 `json:"growth_factor" yaml:"growth_factor"`
}

type Pipeline struct {
	Identifier string     `json:"id" yaml:"id"`
	Functions  []Function `json:"functions" yaml:"functions"`
	Pipes      []Pipe     `json:"pipes" yaml:"pipes"`
	OnStartup  string     `json:"on_startup" yaml:"on_startup,omitempty"`
	OnTeardown string     `json:"on_teardown" yaml:"on_teardown,omitempty"`
}
