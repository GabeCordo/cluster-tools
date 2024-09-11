package scheduler

import (
	"github.com/GabeCordo/cluster-tools/internal/core/components/database"
	"github.com/GabeCordo/cluster-tools/internal/core/interfaces"
	"testing"
)

var testInterval = &interfaces.Interval{
	Minute: 10,
}

var testJob = &interfaces.Job{
	Identifier: "test",
	Module:     "common",
	Cluster:    "vec",
	Config:     "vec",
	Interval:   *testInterval,
	Metadata:   make(map[string]string),
}

var testDupJob = &interfaces.Job{
	Identifier: "test2",
	Module:     "common",
	Cluster:    "vec",
	Config:     "vec",
	Interval:   *testInterval,
	Metadata:   make(map[string]string),
}

var testInterval2 = &interfaces.Interval{
	Minute: 5,
}

var testJob2 = &interfaces.Job{
	Identifier: "test2",
	Module:     "common",
	Cluster:    "vec",
	Config:     "vec",
	Interval:   *testInterval2,
	Metadata:   make(map[string]string),
}

var testJob3 = &interfaces.Job{
	Identifier: "test3",
	Module:     "common",
	Cluster:    "hello",
	Config:     "hello",
	Interval:   *testInterval,
	Metadata:   make(map[string]string),
}

func TestScheduler_Create(t *testing.T) {

	scheduler, err := New(database.NewLocalJobDatabase())
	if err != nil {
		t.Error(err)
		return
	}

	err = scheduler.Jobs.Create(testJob)
	if err != nil {
		t.Error(err)
	}
}

func TestScheduler_GetBy(t *testing.T) {

	scheduler, err := New(database.NewLocalJobDatabase())
	if err != nil {
		t.Error(err)
		return
	}

	// Create 3 Jobs //

	err = scheduler.Jobs.Create(testJob)
	if err != nil {
		t.Error(err)
		return
	}

	err = scheduler.Jobs.Create(testJob2)
	if err != nil {
		t.Error(err)
		return
	}

	err = scheduler.Jobs.Create(testJob3)
	if err != nil {
		t.Error(err)
		return
	}

	// Attempt to Create 1 Dup Job //

	err = scheduler.Jobs.Create(testDupJob)
	if err == nil {
		t.Error("expected testDupJob to be rejected")
		return
	}

	// Attempt to Get All 3 By Module //
	f1 := &interfaces.Filter{Module: "common"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f1); len(foundJobs) != 3 {
		t.Error("expected 3 Jobs to be found for this module")
		return
	}

	// Attempt to Get 2 Jobs By Their Similar Cluster //
	f2 := &interfaces.Filter{Module: "common", Cluster: "vec"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f2); len(foundJobs) != 2 {
		t.Error("expected 2 Jobs to be found with the same module/cluster pair")
		return
	}

	// Attempt to Get 1 Job By Their //
	f3 := &interfaces.Filter{Module: "common", Cluster: "vec", Interval: *testInterval}
	if foundJobs, _ := scheduler.Jobs.GetBy(f3); len(foundJobs) != 1 {
		t.Error("expected 1 job to be found with the module/cluster/interval combo")
	}

	f4 := &interfaces.Filter{Identifier: "test3"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f4); len(foundJobs) != 1 {
		t.Error("expected 1 job to exist with the identifier test3")
	}
}

func TestScheduler_Delete(t *testing.T) {

	scheduler, err := New(database.NewLocalJobDatabase())
	if err != nil {
		t.Error(err)
		return
	}

	// Create 3 Jobs //

	err = scheduler.Jobs.Create(testJob)
	if err != nil {
		t.Error(err)
		return
	}

	err = scheduler.Jobs.Create(testJob2)
	if err != nil {
		t.Error(err)
		return
	}

	err = scheduler.Jobs.Create(testJob3)
	if err != nil {
		t.Error(err)
		return
	}

	// Attempt to Create 1 Dup Job //

	err = scheduler.Jobs.Create(testDupJob)
	if err == nil {
		t.Error("expected testDupJob to be rejected")
		return
	}

	// Attempt to Get All 3 By Module //
	f1 := &interfaces.Filter{Module: "common"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f1); len(foundJobs) != 3 {
		t.Error("expected 3 Jobs to be found for this module")
		return
	}

	// Attempt to Get 2 Jobs By Their Similar Cluster //
	f2 := &interfaces.Filter{Module: "common", Cluster: "vec"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f2); len(foundJobs) != 2 {
		t.Error("expected 2 Jobs to be found with the same module/cluster pair")
		return
	}

	// Attempt to Get 1 Job By Their //
	f3 := &interfaces.Filter{Module: "common", Cluster: "vec", Interval: *testInterval}
	if foundJobs, _ := scheduler.Jobs.GetBy(f3); len(foundJobs) != 1 {
		t.Error("expected 1 job to be found with the module/cluster/interval combo")
	}

	f4 := &interfaces.Filter{Identifier: "test3"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f4); len(foundJobs) != 1 {
		t.Error("expected 1 job to exist with the identifier test3")
	}

	// Delete By Identifier //
	if err = scheduler.Jobs.Delete(f4); err != nil {
		t.Error(err)
		return
	}

	// validate the only common/hello record is deleted //
	f5 := &interfaces.Filter{Module: "common", Cluster: "hello"}
	if foundJobs, _ := scheduler.Jobs.GetBy(f5); len(foundJobs) != 0 {
		t.Error("expected 0 Jobs to exist with the common/hello pair")
		return
	}

	// validate the other records are not affected //
	if foundJobs, _ := scheduler.Jobs.GetBy(f2); len(foundJobs) != 2 {
		t.Error("expected 2 Jobs to be left alone")
		return
	}

	// validate we can delete by the module/cluster/interval pair in f3
	if err = scheduler.Jobs.Delete(f3); err != nil {
		t.Error(err)
		return
	}

	// validate one record remains and it's not the deleted one //
	if foundJobs, _ := scheduler.Jobs.GetBy(f2); len(foundJobs) != 1 {
		t.Error("expected 1 job in common/vec to be left alone")
		return
	} else if foundJobs[0].Identifier != "test2" {
		t.Error("wrong job was deleted")
	}
}
