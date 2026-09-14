package job

import (
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job/in_memory"
)

var testInterval = &database.Interval{
	Minute: 10,
}

var testJob = &job.Job{
	Identifier: "test",
	Namespace:  "common",
	Pipeline:   "vec",
	Interval:   *testInterval,
	Metadata:   make(map[string]string),
}

var testDupJob = &job.Job{
	Identifier: "test2",
	Namespace:  "common",
	Pipeline:   "vec",
	Interval:   *testInterval,
	Metadata:   make(map[string]string),
}

var testInterval2 = &database.Interval{
	Minute: 5,
}

var testJob2 = &job.Job{
	Identifier: "test2",
	Namespace:  "common",
	Pipeline:   "vec",
	Interval:   *testInterval2,
	Metadata:   make(map[string]string),
}

var testJob3 = &job.Job{
	Identifier: "test3",
	Namespace:  "common",
	Pipeline:   "hello",
	Interval:   *testInterval,
	Metadata:   make(map[string]string),
}

func TestScheduler_Create(t *testing.T) {

	scheduler, err := New(in_memory.NewLocalJobDatabase())
	if err != nil {
		t.Error(err)
		return
	}

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob)
	if err != nil {
		t.Error(err)
	}
}

func TestScheduler_GetBy(t *testing.T) {

	scheduler, err := New(in_memory.NewLocalJobDatabase())
	if err != nil {
		t.Error(err)
		return
	}

	// Create 3 Jobs //

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob2)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob3)
	if err != nil {
		t.Error(err)
		return
	}

	// Attempt to Create 1 Dup Job //

	_, err = scheduler.Jobs.Create(database.Filter{}, testDupJob)
	if err == nil {
		t.Error("expected testDupJob to be rejected")
		return
	}

	// Attempt to Get All 3 By Namespace //
	f1 := database.Filter{Namespace: "common"}
	if foundJobs := scheduler.Jobs.Get(f1); len(foundJobs) != 3 {
		t.Error("expected 3 Jobs to be found for this module")
		return
	}

	// Attempt to Get 2 Jobs By Their Similar Function //
	f2 := database.Filter{Namespace: "common", Pipeline: "vec"}
	if foundJobs := scheduler.Jobs.Get(f2); len(foundJobs) != 2 {
		t.Error("expected 2 Jobs to be found with the same module/cluster pair")
		return
	}

	// Attempt to Get 1 Job By Their //
	f3 := database.Filter{Namespace: "common", Pipeline: "vec", Interval: *testInterval}
	if foundJobs := scheduler.Jobs.Get(f3); len(foundJobs) != 1 {
		t.Error("expected 1 job to be found with the module/cluster/interval combo")
	}

	f4 := database.Filter{Identifier: "test3"}
	if foundJobs := scheduler.Jobs.Get(f4); len(foundJobs) != 1 {
		t.Error("expected 1 job to exist with the identifier test3")
	}
}

func TestScheduler_Delete(t *testing.T) {

	scheduler, err := New(in_memory.NewLocalJobDatabase())
	if err != nil {
		t.Error(err)
		return
	}

	// Create 3 Jobs //

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob2)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = scheduler.Jobs.Create(database.Filter{}, testJob3)
	if err != nil {
		t.Error(err)
		return
	}

	// Attempt to Create 1 Dup Job //

	_, err = scheduler.Jobs.Create(database.Filter{}, testDupJob)
	if err == nil {
		t.Error("expected testDupJob to be rejected")
		return
	}

	// Attempt to Get All 3 By Namespace //
	f1 := database.Filter{Namespace: "common"}
	if foundJobs := scheduler.Jobs.Get(f1); len(foundJobs) != 3 {
		t.Error("expected 3 Jobs to be found for this module")
		return
	}

	// Attempt to Get 2 Jobs By Their Similar Function //
	f2 := database.Filter{Namespace: "common", Pipeline: "vec"}
	if foundJobs := scheduler.Jobs.Get(f2); len(foundJobs) != 2 {
		t.Error("expected 2 Jobs to be found with the same module/cluster pair")
		return
	}

	// Attempt to Get 1 Job By Their //
	f3 := database.Filter{Namespace: "common", Pipeline: "vec", Interval: *testInterval}
	if foundJobs := scheduler.Jobs.Get(f3); len(foundJobs) != 1 {
		t.Error("expected 1 job to be found with the module/cluster/interval combo")
	}

	f4 := database.Filter{Identifier: "test3"}
	if foundJobs := scheduler.Jobs.Get(f4); len(foundJobs) != 1 {
		t.Error("expected 1 job to exist with the identifier test3")
	}

	// Delete By Identifier //
	if err = scheduler.Jobs.Delete(f4); err != nil {
		t.Error(err)
		return
	}

	// validate the only common/hello record is deleted //
	f5 := database.Filter{Namespace: "common", Pipeline: "hello"}
	if foundJobs := scheduler.Jobs.Get(f5); len(foundJobs) != 0 {
		t.Error("expected 0 Jobs to exist with the common/hello pair")
		return
	}

	// validate the other records are not affected //
	if foundJobs := scheduler.Jobs.Get(f2); len(foundJobs) != 2 {
		t.Error("expected 2 Jobs to be left alone")
		return
	}

	// validate we can delete by the module/cluster/interval pair in f3
	if err = scheduler.Jobs.Delete(f3); err != nil {
		t.Error(err)
		return
	}

	// validate one record remains and it's not the deleted one //
	if foundJobs := scheduler.Jobs.Get(f2); len(foundJobs) != 1 {
		t.Error("expected 1 job in common/vec to be left alone")
		return
	} else if foundJobs[0].Identifier != "test2" {
		t.Error("wrong job was deleted")
	}
}
