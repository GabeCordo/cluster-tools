package wrapper

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/interfaces"
	"log"
	"sync"
)

type Module struct {
	clusters map[string]*Cluster

	Mounted         bool `json:"mounted"`
	MarkForDeletion bool `json:"mark-for-deletion"`

	Identifier string  `json:"identifier"`
	Version    float64 `json:"version"`

	mutex sync.RWMutex
}

func NewModule() *Module {

	moduleWrapper := new(Module)

	moduleWrapper.clusters = make(map[string]*Cluster)
	moduleWrapper.Mounted = false

	return moduleWrapper
}

func (moduleWrapper *Module) IsMounted() bool {

	return moduleWrapper.Mounted
}

func (moduleWrapper *Module) Mount() *Module {

	moduleWrapper.Mounted = true
	return moduleWrapper
}

func (moduleWrapper *Module) UnMount() *Module {

	moduleWrapper.Mounted = false
	return moduleWrapper
}

func (moduleWrapper *Module) GetClustersData() map[string]bool {

	mounts := make(map[string]bool)

	for identifier, clusterWrapper := range moduleWrapper.clusters {
		mounts[identifier] = clusterWrapper.Mounted
	}

	return mounts
}

func (moduleWrapper *Module) GetClusters() (clusterWrappers []*Cluster) {

	clusterWrappers = make([]*Cluster, 0)

	for _, clusterWrapper := range moduleWrapper.clusters {
		clusterWrappers = append(clusterWrappers, clusterWrapper)
	}

	return clusterWrappers
}

func (moduleWrapper *Module) GetCluster(clusterName string) (clusterWrapper *Cluster, found bool) {

	moduleWrapper.mutex.RLock()
	defer moduleWrapper.mutex.RUnlock()

	clusterWrapper, found = moduleWrapper.clusters[clusterName]
	return clusterWrapper, found
}

// AddCluster
// Creates a new cluster record within the calling module. The cluster name defines the keyword
// an operator uses to provision a cluster, and the mode represents how the cluster is run.
func (moduleWrapper *Module) AddCluster(clusterName string, mode string, implementation cluster.Cluster, cfg ...*cluster.Config) (*Cluster, error) {

	moduleWrapper.mutex.RLock()

	if _, found := moduleWrapper.clusters[clusterName]; found {
		return nil, errors.New("a cluster with this identifier already exists")
	}

	moduleWrapper.mutex.RUnlock()

	moduleWrapper.mutex.Lock()
	defer moduleWrapper.mutex.Unlock()

	clusterWrapper := NewCluster(moduleWrapper.Identifier, clusterName, cluster.EtlMode(mode), implementation)
	clusterWrapper.Mounted = true
	clusterWrapper.DefaultConfig = cluster.DefaultConfig  // copy
	clusterWrapper.DefaultConfig.Identifier = clusterName // copy

	// improve the simplicity of writing quick/simple functions:
	// 1. lower the bar of entry for new developers
	// 2. hide the advanced options when it's not required
	//		- the concept of a 'pipeline' that can be fine-tuned becomes
	//		  important in the optimization phase after many iterations
	if len(cfg) == 0 {
		clusterWrapper.DefaultConfig = cluster.DefaultConfig
		clusterWrapper.DefaultConfig.Identifier = clusterName
	}

	// skip: if the pipeline is never provided since len() == 0
	for _, c := range cfg {
		clusterWrapper.DefaultConfig = *c
		clusterWrapper.DefaultConfig.Identifier = clusterName
	}

	moduleWrapper.clusters[clusterName] = clusterWrapper

	return clusterWrapper, nil
}

func (moduleWrapper *Module) DeleteCluster(identifier string) (deleted, found bool) {

	clusterWrapper, found := moduleWrapper.clusters[identifier]
	if !found {
		return false, false
	}

	if !clusterWrapper.CanDelete() {
		return false, true
	}

	moduleWrapper.mutex.Lock()
	defer moduleWrapper.mutex.Unlock()

	delete(moduleWrapper.clusters, identifier)
	return true, true
}

func (moduleWrapper *Module) CanDelete() (canDelete bool) {

	moduleWrapper.mutex.RLock()
	defer moduleWrapper.mutex.RUnlock()

	// if the module is not marked for deletion, it should not be deleted
	if !moduleWrapper.MarkForDeletion {
		log.Printf("[modules] cannot delete %s - not marked for deletion\n", moduleWrapper.Identifier)
		return false
	}

	canDelete = true
	// look over all the supervisor in a module
	for clusterName, clusterWrapper := range moduleWrapper.clusters {

		if !clusterWrapper.CanDelete() {
			log.Printf("[modules][cluster] cannot delete %s\n", clusterName)
			canDelete = false
			break
		}
	}

	return canDelete
}

func (moduleWrapper *Module) ToConfig() *interfaces.ModuleConfig {
	cfg := new(interfaces.ModuleConfig)

	cfg.Name = moduleWrapper.Identifier
	cfg.Version = moduleWrapper.Version
	cfg.Exports = make([]interfaces.Cluster, len(moduleWrapper.clusters))

	idx := 0
	for _, cluster := range moduleWrapper.clusters {
		cfg.Exports[idx] = interfaces.Cluster{
			Cluster:     cluster.Identifier,
			StaticMount: cluster.Mounted,
			Config: interfaces.ClusterConfig{
				Mode:    interfaces.EtlMode(cluster.Mode),
				OnCrash: interfaces.OnCrash(cluster.DefaultConfig.OnCrash),
				OnLoad:  interfaces.OnLoad(cluster.DefaultConfig.OnLoad),
				Static: struct {
					TFunctions int `yaml:"t-functions" json:"t-functions"`
					LFunctions int `yaml:"l-functions" json:"l-functions"`
				}{
					TFunctions: cluster.DefaultConfig.StartWithNTransformClusters,
					LFunctions: cluster.DefaultConfig.StartWithNTransformClusters,
				},
				Dynamic: struct {
					TFunction interfaces.DynamicFeatures `yaml:"t-function" json:"t-function"`
					LFunction interfaces.DynamicFeatures `yaml:"l-function" json:"l-function"`
				}{
					TFunction: interfaces.DynamicFeatures{
						Threshold:    cluster.DefaultConfig.ETChannelThreshold,
						GrowthFactor: cluster.DefaultConfig.ETChannelGrowthFactor,
					},
					LFunction: interfaces.DynamicFeatures{
						Threshold:    cluster.DefaultConfig.TLChannelThreshold,
						GrowthFactor: cluster.DefaultConfig.TLChannelGrowthFactor,
					},
				},
			},
		}
		idx++
	}

	return cfg
}

func (moduleWrapper *Module) Print() {
	fmt.Printf("%s %.3f\n", moduleWrapper.Identifier, moduleWrapper.Version)
}
