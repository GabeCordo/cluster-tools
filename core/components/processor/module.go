package processor

func (module *Module) add(name string) (success bool) {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.clusters[name]; found {
		return false
	}

	module.clusters[name] = newCluster(name)
	return true
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

func (module *Module) Get(name string) (instance *Cluster, found bool) {

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
		clusters[idx] = clusterInstance.data
		idx++
	}

	return clusters
}
