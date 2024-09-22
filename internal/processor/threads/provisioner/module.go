package provisioner

import "github.com/GabeCordo/cluster-tools/wrapper"

func (thread *Thread) getModules() []*wrapper.Module {

	defer thread.requestWg.Done()
	return GetProvisionerInstance().GetModules()
}
