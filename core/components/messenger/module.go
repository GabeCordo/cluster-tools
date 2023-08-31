package messenger

import "errors"

func (module *Module) Get(identifier string) (*Cluster, bool) {

	module.mutex.RLock()
	defer module.mutex.RUnlock()

	cluster, found := module.clusters[identifier]

	return cluster, found
}

func (module *Module) Create(identifier string) (*Cluster, error) {

	module.mutex.Lock()
	defer module.mutex.Unlock()

	if _, found := module.clusters[identifier]; found {
		return nil, errors.New("cluster already exists")
	}

	cluster := NewCluster()
	module.clusters[identifier] = cluster
	return cluster, nil
}
