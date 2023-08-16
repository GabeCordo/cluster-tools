package processor

func (cluster *Cluster) Mount() {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	cluster.Mounted = true
}

func (cluster *Cluster) Unmount() {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	cluster.Mounted = false
}
