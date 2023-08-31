package processor

import "github.com/GabeCordo/mango/core/interfaces/cluster"

func (c *Cluster) Add(processor *Processor) {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.processors = append(c.processors, processor)
	c.numOfProcessors++
}

func (c *Cluster) IsMounted() bool {
	return c.data.Mounted
}

func (c *Cluster) IsStream() bool {
	return c.data.Mode == cluster.Stream
}

func (c *Cluster) Mount() {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data.Mounted = true
}

func (c *Cluster) Unmount() {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data.Mounted = false
}

func (c *Cluster) SelectProcessor() *Processor {

	// TODO : this is a simple circular shift balancer
	// maybe consider something with the delays the current processors have
	// or number of processes running

	instance := c.processors[c.processorIndex]
	if d := c.numOfProcessors - 1; d != 0 {
		c.processorIndex = (c.processorIndex + 1) % (len(c.processors))
	} else {
		c.processorIndex = 0
	}

	return instance
}
