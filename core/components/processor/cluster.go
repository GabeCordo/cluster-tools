package processor

func (cluster *Cluster) Add(processor *Processor) {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	cluster.Processors = append(cluster.Processors, processor)
}

func (cluster *Cluster) Mount() {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	cluster.data.Mounted = true
}

func (cluster *Cluster) Unmount() {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	cluster.data.Mounted = false
}
