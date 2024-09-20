package provisioner

import "github.com/GabeCordo/clarence/wrapper"

func (thread *Thread) getModules() []*wrapper.Module {

	defer thread.requestWg.Done()
	return GetProvisionerInstance().GetModules()
}
