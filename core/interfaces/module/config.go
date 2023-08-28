package module

import "github.com/GabeCordo/mango/core/interfaces/cluster"

type DynamicFeatures struct {
	Threshold    int `yaml:"threshold" json:"threshold"`
	GrowthFactor int `yaml:"growth-factor" json:"growth-factor"`
}

type ClusterConfig struct {
	Mode    cluster.EtlMode `yaml:"mode" json:"mode"`
	OnCrash cluster.OnCrash `yaml:"on-crash" json:"on-crash"`
	OnLoad  cluster.OnLoad  `yaml:"on-load" json:"on-load"`
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

type Config struct {
	Name    string    `yaml:"name" json:"name"`
	Version float64   `yaml:"version" json:"version"`
	Contact Contact   `yaml:"contact,omitempty" json:"contact,omitempty"`
	Exports []Cluster `yaml:"exports" json:"exports"`
}

func (c Cluster) ToClusterConfig() cluster.Config {

	return cluster.Config{
		Identifier:                  c.Cluster,
		OnLoad:                      c.Config.OnLoad,
		OnCrash:                     c.Config.OnCrash,
		StartWithNTransformClusters: c.Config.Static.TFunctions,
		StartWithNLoadClusters:      c.Config.Static.LFunctions,
		ETChannelThreshold:          c.Config.Dynamic.TFunction.Threshold,
		ETChannelGrowthFactor:       c.Config.Dynamic.TFunction.GrowthFactor,
		TLChannelThreshold:          c.Config.Dynamic.LFunction.Threshold,
		TLChannelGrowthFactor:       c.Config.Dynamic.LFunction.GrowthFactor,
	}
}
