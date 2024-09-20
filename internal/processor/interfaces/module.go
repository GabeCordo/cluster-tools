package interfaces

type DynamicFeatures struct {
	Threshold    int     `yaml:"threshold" json:"threshold"`
	GrowthFactor float64 `yaml:"growth-factor" json:"growth-factor"`
}

type ClusterConfig struct {
	Mode    EtlMode `yaml:"mode" json:"mode"`
	OnCrash OnCrash `yaml:"on-crash" json:"on-crash"`
	OnLoad  OnLoad  `yaml:"on-load" json:"on-load"`
	Static  struct {
		TFunctions int `yaml:"t-functions" json:"t-functions"`
		LFunctions int `yaml:"l-functions" json:"l-functions"`
	} `yaml:"static"`
	Dynamic struct {
		TFunction DynamicFeatures `yaml:"t-function" json:"t-function"`
		LFunction DynamicFeatures `yaml:"l-function" json:"l-function"`
	} `yaml:"dynamic"`
}

type Cluster struct {
	Cluster     string        `yaml:"cluster" json:"cluster"`
	StaticMount bool          `yaml:"mount" json:"mount"`
	Config      ClusterConfig `yaml:"config" json:"config"`
}

type Contact struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type ModuleConfig struct {
	Name    string    `yaml:"name" json:"name"`
	Version float64   `yaml:"version" json:"version"`
	Contact Contact   `yaml:"contact,omitempty" json:"contact,omitempty"`
	Exports []Cluster `yaml:"exports" json:"clusters"`
}
