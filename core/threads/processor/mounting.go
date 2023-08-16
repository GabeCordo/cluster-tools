package processor

func (thread *Thread) mountModule(name string) {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return
	}

	instance.Mount()
}

func (thread *Thread) unmountModule(name string) {

	instance, found := GetTableInstance().Get(name)
	if !found {
		return
	}

	instance.Unmount()
}

func (thread *Thread) mountCluster(moduleName, clusterName string) {

	moduleInstance, found := GetTableInstance().Get(moduleName)
	if !found {
		return
	}

	clusterInstance, found := moduleInstance.Get(clusterName)
	if !found {
		return
	}

	clusterInstance.Mount()
}

func (thread *Thread) unmountCluster(moduleName, clusterName string) {

	moduleInstance, found := GetTableInstance().Get(moduleName)
	if !found {
		return
	}

	clusterInstance, found := moduleInstance.Get(clusterName)
	if !found {
		return
	}

	clusterInstance.Unmount()
}
