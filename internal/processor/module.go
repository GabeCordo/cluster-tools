package processor

import (
	"github.com/GabeCordo/cluster-tools/internal/database/config"
	"gopkg.in/yaml.v3"
	"os"
)

type DynamicFeatures struct {
	Threshold    int     `yaml:"threshold" json:"threshold"`
	GrowthFactor float64 `yaml:"growth-factor" json:"growth-factor"`
}

type ModuleClusterConfig struct {
	Mode    config.EtlMode `yaml:"mode" json:"mode"`
	OnCrash config.OnCrash `yaml:"on-crash" json:"on-crash"`
	OnLoad  config.OnLoad  `yaml:"on-load" json:"on-load"`
	Static  struct {
		TFunctions int `yaml:"t-functions" json:"t-functions"`
		LFunctions int `yaml:"l-functions" json:"l-functions"`
	} `yaml:"static"`
	Dynamic struct {
		TFunction DynamicFeatures `yaml:"t-function" json:"t-function"`
		LFunction DynamicFeatures `yaml:"l-function" json:"l-function"`
	} `yaml:"dynamic"`
}

type ModuleCluster struct {
	Cluster     string              `yaml:"cluster" json:"cluster"`
	StaticMount bool                `yaml:"mount" json:"mount"`
	Config      ModuleClusterConfig `yaml:"config" json:"config"`
}

type ModuleContact struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type ModuleConfig struct {
	Name    string          `yaml:"name" json:"name"`
	Version float64         `yaml:"version" json:"version"`
	Contact ModuleContact   `yaml:"contact,omitempty" json:"contact,omitempty"`
	Exports []ModuleCluster `yaml:"exports" json:"clusters"`
}

func (c ModuleCluster) ToClusterConfig() config.Config {

	return config.Config{
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

func ConfigFromYAML(path string) (*ModuleConfig, error) {

	config := new(ModuleConfig)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(bytes, config)

	return config, nil
}

func (config ModuleConfig) Verify() bool {

	// ensure that every export identifier is unique
	exports := make(map[string]bool)
	for _, export := range config.Exports {
		if _, found := exports[export.Cluster]; found {
			return false
		} else {
			exports[export.Cluster] = true
		}
	}

	return true
}

func (module *Module) addCluster(name string) (success bool) {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.clusters[name]; found {
		return false
	}

	module.clusters[name] = newCluster(name)
	return true
}

func (module *Module) IsMounted() bool {
	return module.data.Mounted
}

func (module *Module) Mount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.data.Mounted = true
}

func (module *Module) Unmount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.data.Mounted = false
}

func (module *Module) GetData() ModuleData {

	return module.data
}

func (module *Module) GetCluster(name string) (instance *Cluster, found bool) {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	instance, found = module.clusters[name]
	return instance, found
}

func (module *Module) Registered() []ClusterData {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	clusters := make([]ClusterData, len(module.clusters))

	idx := 0
	for _, clusterInstance := range module.clusters {
		clusters[idx] = clusterInstance.GetData()
		idx++
	}

	return clusters
}
