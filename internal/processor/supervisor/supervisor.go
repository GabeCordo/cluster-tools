package supervisor

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/components/channel"
	internal_cluster "github.com/GabeCordo/cluster-tools/internal/processor/components/cluster"
	"log"
	"time"
)

const (
	DefaultNumberOfClusters       = 1
	DefaultMonitorRefreshDuration = 100
	DefaultChannelThreshold       = 10
	DefaultChannelGrowthFactor    = 2
)

func (supervisor *Supervisor) Event(event Event) bool {
	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	if supervisor.State == UnTouched {
		if event == Startup {
			supervisor.State = Running
		} else if (event == Suspend) || (event == TearedDown) {
			supervisor.State = Stopping
		} else {
			return false
		}
	} else if supervisor.State == Running {
		if event == StartProvision {
			supervisor.State = Provisioning
		} else if event == Error {
			supervisor.State = Failed
		} else if event == Suspend {
			supervisor.State = Stopping
		} else if event == TearedDown {
			supervisor.State = Terminated
		} else {
			return false
		}
	} else if supervisor.State == Provisioning {
		if event == EndProvision {
			supervisor.State = Running
		} else if event == Error {
			supervisor.State = Failed
		} else if event == Suspend {
			supervisor.State = Stopping
		} else {
			return false
		}
	} else if supervisor.State == Stopping {
		if event == TearedDown {
			supervisor.State = Terminated
		} else {
			return false
		}
	} else if (supervisor.State == Failed) || (supervisor.State == Terminated) {
		return false
	}

	return true // represents a boolean ~ hasStateChanged?
}

func (supervisor *Supervisor) IsAlive() bool {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return (supervisor.State != Failed) && (supervisor.State != Terminated)
}

func (supervisor *Supervisor) Start() (response *internal_cluster.Response) {
	supervisor.Event(Startup)

	defer supervisor.Event(TearedDown)

	defer func() {
		// has the user defined function crashed during runtime?
		if r := recover(); r != nil {
			// yes => return a response that identifies that the cluster crashed
			response = internal_cluster.NewResponse(
				supervisor.Config,
				supervisor.Stats,
				time.Now().Sub(supervisor.StartTime),
				true,
			)
		}
	}()

	supervisor.StartTime = time.Now()

	if sysFunc, ok := (supervisor.group).(internal_cluster.SystemFunctions); ok {
		sysFunc.Setup(supervisor.StartTime, supervisor.helper)
	}

	// the common specifies the number of load functions to quitE running in parallel
	for i := 0; i < supervisor.Config.StartWithNLoadClusters; i++ {
		supervisor.Provision(internal_cluster.Load)
		supervisor.Stats.Threads.NumActiveLoadRoutines++
		supervisor.Stats.Threads.NumProvisionedLoadRoutines++
		supervisor.ActiveLRoutines++
	}

	// the common specifies the number of transform functions to quitE running in parallel
	for i := 0; i < supervisor.Config.StartWithNTransformClusters; i++ {
		supervisor.Provision(internal_cluster.Transform)
		supervisor.Stats.Threads.NumActiveTransformRoutines++
		supervisor.Stats.Threads.NumProvisionedTransformRoutes++
		supervisor.ActiveTRoutines++
	}

	//// quitE creating the default frontend goroutines
	supervisor.Provision(internal_cluster.Extract)
	supervisor.Stats.Threads.NumProvisionedExtractRoutines++
	supervisor.Stats.Threads.NumActiveExtractRoutines++

	//// end creating the default frontend goroutines

	// every N seconds we should check if the ETChannel or TLChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	// calculate the timings produced by data being fed across each of the channels
	supervisor.CalculateTiming()

	response = internal_cluster.NewResponse(
		supervisor.Config,
		supervisor.Stats,
		time.Now().Sub(supervisor.StartTime),
		false,
	)

	return response
}

func (supervisor *Supervisor) Teardown() {

	if sysFunc, ok := (supervisor.group).(internal_cluster.SystemFunctions); ok {
		sysFunc.Teardown(time.Now(), supervisor.helper)
	}

	supervisor.Event(Suspend)
}

func (supervisor *Supervisor) Runtime() {
	for {
		if supervisor.State == Terminated {
			break
		}

		etChannelState := supervisor.ETChannel.GetState()

		if (supervisor.State == Stopping) && supervisor.ETChannel.Accepting() {
			supervisor.ETChannel.StopPushes()
		}

		if etChannelState == channel.Congested {
			// when the ET channel is congested provision new Transform
			// functions in-accordance to the ET growth-factor
			supervisor.Stats.Channels.NumEtThresholdBreaches++
			n := supervisor.ETChannel.Config.GrowthFactor
			for n > 0 {
				supervisor.Stats.Threads.NumProvisionedTransformRoutes++
				supervisor.Stats.Threads.NumActiveTransformRoutines++
				supervisor.Provision(internal_cluster.Transform)
				supervisor.ActiveTRoutines++
				n--
			}
		} else if (etChannelState == channel.Underutilized) || (etChannelState == channel.Idle) {
			n := supervisor.ETChannel.Config.GrowthFactor
			for n > 0 {
				// never remove all transform nodes otherwise we risk the
				// ET channel having no consumers
				if supervisor.ActiveTRoutines <= 1 {
					break
				}
				supervisor.ActiveTRoutines--
				supervisor.Stats.Threads.NumActiveTransformRoutines--
				supervisor.Remove(internal_cluster.Transform)
				n--
			}
		}

		tlChannelState := supervisor.TLChannel.GetState()

		if tlChannelState == channel.Congested {
			// when the TL channel is congested provision new Load
			// functions in-accordance to the TL growth-factor
			supervisor.Stats.Channels.NumTlThresholdBreaches++
			n := supervisor.TLChannel.Config.GrowthFactor
			for n > 0 {
				supervisor.Stats.Threads.NumProvisionedLoadRoutines++
				supervisor.Stats.Threads.NumActiveLoadRoutines++
				supervisor.Provision(internal_cluster.Load)
				supervisor.ActiveLRoutines++
				n--
			}
		} else if (tlChannelState == channel.Underutilized) || (tlChannelState == channel.Idle) {
			n := supervisor.TLChannel.Config.GrowthFactor
			for n > 0 {
				// never remove all transform nodes otherwise we risk the
				// ET channel having no consumers
				if supervisor.ActiveLRoutines <= 1 {
					break
				}
				supervisor.ActiveLRoutines--
				supervisor.Stats.Threads.NumActiveLoadRoutines--
				supervisor.Remove(internal_cluster.Load)
				n--
			}
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Millisecond)
	}
}

func (supervisor *Supervisor) ExtractWrapper(h cluster.H, m cluster.M, out cluster.Out) <-chan struct{} {
	done := make(chan struct{})

	// the function always finishes till completion unless a direct shutdown is called on the server
	// which stops data collection from some source
	go func() {
		defer func() {
			done <- struct{}{}
			close(done)
		}()
		supervisor.group.ExtractFunc(h, m, out)
	}()
	return done
}

func (supervisor *Supervisor) ExtractShutdownWrapper() <-chan struct{} {
	done := make(chan struct{})

	// we need to create a separate goroutine otherwise it will block the current
	// thread from re-evaluating the select statement wherever the ExtractShutdownWrapper is called
	go func() {
		defer close(done)
		for {
			// the IsAlive clause ensures that once a supervisor is dead, we will not leak memory
			// with a forever-running goroutine
			if (supervisor.State == Stopping) || (!supervisor.IsAlive()) {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return done
}

func (supervisor *Supervisor) Provision(segment internal_cluster.Segment) {
	supervisor.Event(StartProvision)
	defer supervisor.Event(EndProvision)

	// the supervisor must provision new threads one at a time.
	// note: avoid the possibility of >1 thread modifying the wait-group at a time
	supervisor.threadMutex.Lock()
	defer supervisor.threadMutex.Unlock()

	var quit chan bool = make(chan bool)
	switch segment {
	case internal_cluster.Transform:
		supervisor.quitT = append(supervisor.quitT, quit)
	case internal_cluster.Load:
		supervisor.quitL = append(supervisor.quitL, quit)
	default:
		quit = nil
	}

	go func(supervisor *Supervisor, quit chan bool) {
		switch segment {
		case internal_cluster.Extract:
			{
				defer func() {
					if r := recover(); r != nil {
						log.Println(r)
						log.Println("cluster.Extract function raised error")
						supervisor.ETChannel.ProducerDone()
						supervisor.waitGroup.Done()
					}
				}()

				oneWayChannel, _ := internal_cluster.NewOneWayManagedChannel(supervisor.ETChannel)
				supervisor.ETChannel.AddProducer()

				select {
				case <-supervisor.ExtractWrapper(supervisor.helper, supervisor.Metadata, oneWayChannel):
					break
				case <-supervisor.ExtractShutdownWrapper():
					fmt.Println("shutdown caused extract to finish early")
					break
				}

				// if the number of producers is 0, the ET channel will close that
				// allows the Transform goroutines to terminate once they have
				// completed processing all of their data
				supervisor.ETChannel.ProducerDone()
			}
		case internal_cluster.Transform:
			{
				defer func() {
					if r := recover(); r != nil {
						log.Println("cluster.Transform function raised error")
						log.Println(r)
						supervisor.TLChannel.ProducerDone()
						supervisor.waitGroup.Done()
					}
				}()

				supervisor.TLChannel.AddProducer()
				closeChan := false

				for {
					select {
					case request := <-supervisor.ETChannel.GetChannel():
						{
							// sometimes we are receiving bad data?
							if request.IsInvalid() {
								closeChan = true
								break
							}

							// associates a TimeOut to the data being removed from the channel and decrements
							// the data counter for the current pipe
							supervisor.ETChannel.DataPopped(request.In)

							supervisor.mutexET.Lock()
							supervisor.Stats.Data.TotalOverETChannel++
							supervisor.mutexET.Unlock()

							if i, ok := (supervisor.group).(internal_cluster.VerifiableET); ok && !i.VerifyETFunction(request) {
								continue
							}

							data, success := supervisor.group.TransformFunc(supervisor.helper, supervisor.Metadata, request.Data)
							if success {
								supervisor.mutexTL.Lock()
								supervisor.Stats.Data.TotalOverTLChannel++
								supervisor.mutexTL.Unlock()
								if data == nil {
									fmt.Println("data is nil")
								}
								supervisor.TLChannel.Push(data)
							}
						}
					case <-quit:
						{
							closeChan = true
						}
					}

					if closeChan {
						break
					}
				}

				// if the number of producers is 0, the TL channel will close that
				// allows the Load goroutines to terminate once they have
				// completed processing all of their data
				supervisor.TLChannel.ProducerDone()
			}
		case internal_cluster.Load:
			{
				defer func() {
					if r := recover(); r != nil {
						log.Println(r)
						log.Println("cluster.Load function raised error")
						supervisor.waitGroup.Done()
					}
				}()

				aggregatedData := make([]any, 0)
				closeChan := false

				for {
					select {
					case request := <-supervisor.TLChannel.GetChannel():
						{
							if request.IsInvalid() {
								closeChan = true
								break
							}

							supervisor.mutexTL.Lock()
							supervisor.Stats.Data.TotalProcessed++
							supervisor.mutexTL.Unlock()

							// associates a TimeOut to the data being removed from the channel and decrements
							// the data counter for the current pipe
							supervisor.TLChannel.DataPopped(request.In)

							if i, ok := (supervisor.group).(internal_cluster.VerifiableTL); ok && !i.VerifyTLFunction(request) {
								continue
							}

							if a, clusterUsesLoadOneByOne := (supervisor.group).(internal_cluster.LoadOne); clusterUsesLoadOneByOne {
								a.LoadFunc(supervisor.helper, supervisor.Metadata, request.Data)
							} else if _, success := (supervisor.group).(internal_cluster.LoadAll); success {
								aggregatedData = append(aggregatedData, request.Data)
							}
						}
					case <-quit:
						{
							closeChan = true
						}
					}

					if closeChan {
						break
					}
				}

				if a, clusterUsesLoadAllAtEnd := (supervisor.group).(internal_cluster.LoadAll); clusterUsesLoadAllAtEnd {
					a.LoadFunc(supervisor.helper, supervisor.Metadata, aggregatedData)
				}
			}
		default:
			{
				panic("provisioning invalid cluster type")
			}
		}

		// notify the wait group a process has completed ~ if all are finished we close the monitor
		supervisor.waitGroup.Done()
	}(supervisor, quit)

	// a new function (E, T, or L) is provisioned
	// we should inform the wait group that the supervisor isn't finished until the wg is done
	supervisor.waitGroup.Add(1)
}

func (supervisor *Supervisor) Remove(segment internal_cluster.Segment) {

	switch segment {
	case internal_cluster.Transform:
		{
			if supervisor.ActiveTRoutines <= 0 {
				panic("attempting to quitE when no transform functions are running")
			}
			quit := supervisor.quitT[0]
			supervisor.quitT = supervisor.quitT[1:]
			quit <- true
		}
	case internal_cluster.Load:
		{
			if supervisor.ActiveLRoutines <= 0 {
				panic("attempting to quite when no load functions are running")
			}
			quit := supervisor.quitL[0]
			supervisor.quitL = supervisor.quitL[1:]
			quit <- true
		}
	default:
		{
			panic("removing invalid cluster function type")
		}
	}
}

func (supervisor *Supervisor) Deletable() bool {
	return (supervisor.State == Terminated) || (supervisor.State == Failed)
}

func (supervisor *Supervisor) CalculateTiming() {

	supervisor.Stats.Data.TotalDropped = supervisor.Stats.Data.TotalOverETChannel - supervisor.Stats.Data.TotalOverTLChannel
	supervisor.Stats.CalculateTiming(supervisor.ETChannel.Statistics, supervisor.TLChannel.Statistics)
}

func (supervisor *Supervisor) Print() {
	fmt.Printf("Id: %d\n", supervisor.Id)
	fmt.Printf("Cluster: %s\n", supervisor.Config.Identifier)
}

func (status Status) ToString() string {
	switch status {
	case UnTouched:
		return "UnTouched"
	case Running:
		return "Running"
	case Provisioning:
		return "Provisioning"
	case Failed:
		return "Failed"
	case Terminated:
		return "Terminated"
	default:
		return "None"
	}
}
