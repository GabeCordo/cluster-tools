package provision

import (
	"fmt"
	"sync"

	"github.com/GabeCordo/FunctionScheduler/internal/shared/buffers"
)

type Provisioner struct {
	repository *ScalingFunctions.Repository
	runs       map[uint64]ScalingFunctions.Interactable
	mutex      sync.RWMutex
	activeRuns uint64
	runIds     *buffers.RingBuffer
}

func New(repository *ScalingFunctions.Repository, ids *buffers.RingBuffer) *Provisioner {

	provisioner := new(Provisioner)
	if provisioner == nil {
		panic("failed to allocate memory for Provisioner struct")
	}
	provisioner.runs = make(map[uint64]ScalingFunctions.Interactable)
	provisioner.runIds = ids
	provisioner.repository = repository
	provisioner.activeRuns = 0

	return provisioner
}

func (provisioner *Provisioner) NumberOfActiveSupervisors() uint64 {

	return provisioner.activeRuns
}

func (provisioner *Provisioner) RunExists(id uint64) bool {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	_, found := provisioner.runs[id]
	return found
}

func (provisioner *Provisioner) CreateRun(namespace string, identifier uint64, metadata map[string]string, core string, pipeline *ScalingFunctions.PipelineIR) (ScalingFunctions.Interactable, error) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	p := ScalingFunctions.Build(pipeline, provisioner.repository)
	i := p.Interactable()
	i.Id = identifier

	provisioner.activeRuns++
	provisioner.runs[identifier] = i

	return i, nil
}

func (provisioner *Provisioner) DeleteRun(id uint64) (deleted, found bool) {

	provisioner.mutex.RLock()

	runInstance, found := provisioner.runs[id]
	if !found {
		return false, false
	}

	provisioner.mutex.RUnlock()

	found = true
	if !runInstance.IsRunning() {
		provisioner.mutex.Lock()
		defer provisioner.mutex.Unlock()
		delete(provisioner.runs, id)
		provisioner.activeRuns--
		deleted = true
	} else {
		deleted = false
	}

	return deleted, found
}

func (provisioner *Provisioner) GetRun(id uint64) (ScalingFunctions.Interactable, bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	if r, found := provisioner.runs[id]; found {
		return r, true
	} else {
		return ScalingFunctions.Interactable{}, false
	}
}

func (provisioner *Provisioner) GetRuns() (runs []ScalingFunctions.Interactable) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	runs = make([]ScalingFunctions.Interactable, provisioner.activeRuns)
	for _, r := range provisioner.runs {
		runs = append(runs, r)
	}
	return runs
}

func (provisioner *Provisioner) SuspendRuns() {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	for id, r := range provisioner.runs {
		fmt.Printf("tearing down run %d\n", id)
		r.Stop()
	}
}
