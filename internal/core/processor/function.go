package processor

import (
	"sync"
)

type FunctionData struct {
	Name       string
	Mounted    bool
	Parameters []string
	Returns    []string
}

type Function struct {
	data FunctionData

	processors      []*Processor
	numOfProcessors int
	processorIndex  int

	mutex sync.Mutex
}

func newFunction(builder *ModuleFunction) *Function {
	function := new(Function)

	function.data.Name = builder.Name
	function.data.Parameters = make([]string, len(builder.Parameters))
	copy(function.data.Parameters, builder.Parameters)
	function.data.Returns = make([]string, len(builder.Returns))
	copy(function.data.Returns, builder.Returns)
	function.data.Mounted = false
	function.processors = make([]*Processor, 0)
	function.processorIndex = 0

	return function
}

func (c *Function) Add(processor *Processor) {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.processors = append(c.processors, processor)
	c.numOfProcessors++
}

func (c *Function) IsMounted() bool {
	return c.data.Mounted
}

func (c *Function) Mount() {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data.Mounted = true
}

func (c *Function) Unmount() {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data.Mounted = false
}

func (c *Function) GetData() FunctionData {
	return c.data
}

func (c *Function) SelectProcessor() *Processor {

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
