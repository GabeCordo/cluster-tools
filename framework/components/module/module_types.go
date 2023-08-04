package module

import (
	"github.com/GabeCordo/etl-light/components/cluster"
	"plugin"
)

type DynamicFeatures struct {
	Threshold    int `yaml:"threshold"`
	GrowthFactor int `yaml:"growth-factor"`
}

type ClusterConfig struct {
	Mode    cluster.EtlMode `yaml:"mode"`
	OnCrash cluster.OnCrash `yaml:"on-crash"`
	OnLoad  cluster.OnLoad  `yaml:"on-load"`
	Static  struct {
		TFunctions int `yaml:"t-functions"`
		LFunctions int `yaml:"l-functions"`
	} `yaml:"static"`
	Dynamic struct {
		TFunction DynamicFeatures `yaml:"t-function"`
		LFunction DynamicFeatures `yaml:"l-function"`
	} `yaml:"dynamic"`
}

type Cluster struct {
	Cluster     string        `yaml:"cluster"`
	StaticMount bool          `yaml:"mount"`
	Config      ClusterConfig `yaml:"config"`
}

type Contact struct {
	Name  string `yaml:"name,omitempty"`
	Email string `yaml:"email,omitempty"`
}

type Config struct {
	Identifier string    `yaml:"identifier,omitempty"`
	Version    float64   `yaml:"version,omitempty"`
	Contact    Contact   `yaml:"contact,omitempty"`
	Exports    []Cluster `yaml:"exports"`
}

type Module struct {
	Plugin *plugin.Plugin
	Config *Config
}

type RemoteModule struct {
	Path string
}
