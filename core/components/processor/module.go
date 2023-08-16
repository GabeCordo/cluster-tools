package processor

func (module *Module) add(name string) (success bool) {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.Clusters[name]; found {
		return false
	}

	module.Clusters[name] = newCluster(name)
	return true
}

func (module *Module) Mount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	module.Mounted = true
}

func (module *Module) Unmount() {

	module.mutex.Lock()
	defer module.mutex.Unlock()
	
	module.Mounted = false
}

func (module *Module) Get(name string) (instance *Cluster, found bool) {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	instance, found = module.Clusters[name]
	return instance, found
}

func (module *Module) Registered() map[string]bool {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	clusters := make(map[string]bool)

	for identifier, clusterInstance := range module.Clusters {
		clusters[identifier] = clusterInstance.Mounted
	}

	return clusters
}
