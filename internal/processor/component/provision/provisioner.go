package provision

import (
	"fmt"
	"github.com/GabeCordo/Flock/internal/shared/buffers"
	"github.com/GabeCordo/plover"
	"sync"
)

type Provisioner struct {
	repository *plover.Repository
	runs       map[uint64]plover.Interactable
	mutex      sync.RWMutex
	activeRuns uint64
	runIds     *buffers.RingBuffer
}

func New(repository *plover.Repository, ids *buffers.RingBuffer) *Provisioner {

	provisioner := new(Provisioner)
	if provisioner == nil {
		panic("failed to allocate memory for Provisioner struct")
	}
	provisioner.runs = make(map[uint64]plover.Interactable)
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

func (provisioner *Provisioner) CreateRun(namespace string, identifier uint64, metadata map[string]string, core string, pipeline *plover.PipelineIR) (plover.Interactable, error) {

	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	p := plover.Build(pipeline, provisioner.repository)
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

func (provisioner *Provisioner) GetRun(id uint64) (plover.Interactable, bool) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	if r, found := provisioner.runs[id]; found {
		return r, true
	} else {
		return plover.Interactable{}, false
	}
}

func (provisioner *Provisioner) GetRuns() (runs []plover.Interactable) {
	provisioner.mutex.RLock()
	defer provisioner.mutex.RUnlock()

	runs = make([]plover.Interactable, provisioner.activeRuns)
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
