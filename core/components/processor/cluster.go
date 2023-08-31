package processor

func (cluster *Cluster) Add(processor *Processor) {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	cluster.processors = append(cluster.processors, processor)
	cluster.numOfProcessors++
}

func (cluster *Cluster) IsMounted() bool {
	return cluster.data.Mounted
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

func (cluster *Cluster) SelectProcessor() *Processor {

	// TODO : this is a simple circular shift balancer
	// maybe consider something with the delays the current processors have
	// or number of processes running

	instance := cluster.processors[cluster.processorIndex]
	if d := cluster.numOfProcessors - 1; d != 0 {
		cluster.processorIndex = (cluster.processorIndex + 1) % (len(cluster.processors))
	} else {
		cluster.processorIndex = 0
	}

	return instance
}
