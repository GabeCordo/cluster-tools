package supervisor

import (
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/processor/channel/duplex"
	"log"
	"reflect"
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

func (supervisor *Supervisor) Start() (response *Response) {
	supervisor.Event(Startup)

	defer supervisor.Event(TearedDown)

	defer func() {
		// has the user defined function crashed during runtime?
		if r := recover(); r != nil {
			// yes => return a response that identifies that the cluster crashed
			response = NewResponse(
				supervisor.Stats,
				time.Now().Sub(supervisor.StartTime),
				true,
			)
		}
	}()

	supervisor.StartTime = time.Now()

	//if sysFunc, ok := (supervisor.group).(SystemFunctions); ok {
	//	sysFunc.Setup(supervisor.StartTime, supervisor.helper)
	//}

	for i, function := range supervisor.Pipeline.Functions {
		for j := 0; j < function.StartWith; j++ {
			supervisor.Provision(i)
			supervisor.Stats.Functions[i].Active++
			supervisor.Stats.Functions[i].Provisions++
		}
	}

	// the common specifies the number of load functions to quitE running in parallel
	//for i := 0; i < supervisor.Config.StartWithNLoadClusters; i++ {
	//	supervisor.Provision(Load)
	//	supervisor.Stats.Threads.NumActiveLoadRoutines++
	//	supervisor.Stats.Threads.NumProvisionedLoadRoutines++
	//	supervisor.ActiveLRoutines++
	//}

	// the common specifies the number of transform functions to quitE running in parallel
	//for i := 0; i < supervisor.Config.StartWithNTransformClusters; i++ {
	//	supervisor.Provision(Transform)
	//	supervisor.Stats.Threads.NumActiveTransformRoutines++
	//	supervisor.Stats.Threads.NumProvisionedTransformRoutes++
	//	supervisor.ActiveTRoutines++
	//}

	//// quitE creating the default frontend goroutines
	//supervisor.Provision(Extract)
	//supervisor.Stats.Threads.NumProvisionedExtractRoutines++
	//supervisor.Stats.Threads.NumActiveExtractRoutines++

	//// end creating the default frontend goroutines

	// every N seconds we should check if the ETChannel or TLChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	// calculate the timings produced by data being fed across each of the channels
	// TODO: support
	//supervisor.CalculateTiming()

	response = NewResponse(
		supervisor.Stats,
		time.Now().Sub(supervisor.StartTime),
		false,
	)

	return response
}

func (supervisor *Supervisor) Teardown() {

	//if sysFunc, ok := (supervisor.group).(SystemFunctions); ok {
	//	sysFunc.Teardown(time.Now(), supervisor.helper)
	//}

	supervisor.Event(Suspend)
}

func (supervisor *Supervisor) Runtime() {
	for {
		if supervisor.State == Terminated {
			break
		}

		for i, channel := range supervisor.channels {

			channelState := channel.GetState()

			if (supervisor.State == Stopping) && channel.Accepting() {
				channel.StopPushes()
			}

			if channelState == duplex.Congested {

				supervisor.Stats.Pipes[i].Breaches++
				n := channel.GetGrowthFactor()
				for n > 0 {
					supervisor.Stats.Functions[i+1].Provisions++
					supervisor.Stats.Functions[i+1].Active++
					supervisor.Provision(i + 1)
					n--
				}
			} else if (channelState == duplex.Underutilized) || (channelState == duplex.Idle) {
				n := channel.GetGrowthFactor()
				for n > 0 {
					// never remove all transform nodes otherwise we risk the
					// ET channel having no consumers
					if supervisor.Stats.Functions[i+1].Active <= 1 {
						break
					}
					supervisor.Stats.Functions[i+1].Active--
					supervisor.RemoveProducerFrom(i + 1)
					n--
				}
			}
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Millisecond)
	}
}

// func (supervisor *Supervisor) ExtractWrapper(h cluster.H, m cluster.M, out cluster.Out) <-chan struct{} {
func (supervisor *Supervisor) ExtractWrapper(function any, channel *duplex.ManagedChannel) <-chan struct{} {
	done := make(chan struct{})

	// the function always finishes till completion unless a direct shutdown is called on the server
	// which stops data collection from some source
	go func() {
		defer func() {
			done <- struct{}{}
			close(done)
		}()

		arguments := make([]reflect.Value, 0)

		if reflect.TypeOf(function).NumIn() > 0 {

			// do we expect to pass a pipe?
			channelType := reflect.TypeOf(function).In(0)

			if channelType.Kind() == reflect.Chan {
				channel := reflect.MakeChan(channelType.Elem(), 0)
				arguments = append(arguments, channel)
			}
		}

		go reflect.ValueOf(function).Call(arguments)

		if reflect.TypeOf(function).NumIn() > 0 {

			for {
				value, ok := arguments[0].Recv()
				if !ok {
					channel.Push([]reflect.Value{value})
				}
			}
		}
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
			// the IsAlive clause ensures that once a runner is dead, we will not leak memory
			// with a forever-running goroutine
			if (supervisor.State == Stopping) || (!supervisor.IsAlive()) {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return done
}

func (supervisor *Supervisor) Provision(functionInstance int) {
	supervisor.Event(StartProvision)
	defer supervisor.Event(EndProvision)

	// the runner must provision new threads one at a time.
	// note: avoid the possibility of >1 thread modifying the wait-group at a time
	supervisor.threadMutex.Lock()
	defer supervisor.threadMutex.Unlock()

	var from *duplex.ManagedChannel = nil
	var fromIdx int = -1
	var quit chan bool = nil
	if supervisor.Pipeline.Functions[functionInstance].From != "" {
		for i, c := range supervisor.channels {
			if c.Name == supervisor.Pipeline.Functions[functionInstance].From {
				from = c
				fromIdx = i

				// create a new channel to tell this function to stop listening for data
				quit = make(chan bool)
				supervisor.quit[i] = quit
				return
			}
		}
	}

	var to *duplex.ManagedChannel = nil
	var toIdx int = -1
	if supervisor.Pipeline.Functions[functionInstance].To != "" {
		for i, c := range supervisor.channels {
			if c.Name == supervisor.Pipeline.Functions[functionInstance].To {
				to = c
				toIdx = i
				return
			}
		}
	}

	go func(supervisor *Supervisor, from *duplex.ManagedChannel, fromIdx int, to *duplex.ManagedChannel, toIdx int, quit chan bool) {

		if (from == nil) && (to != nil) {
			// the function is a STARTING NODE of the pipeline if no data is being received
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
					log.Println("cluster.Extract function raised error")
					to.ProducerDone()
					supervisor.waitGroup.Done()
				}
			}()

			select {
			case <-supervisor.ExtractWrapper(supervisor.functions[functionInstance], to):
				break
			case <-supervisor.ExtractShutdownWrapper():
				fmt.Println("shutdown caused extract to finish early")
				break
			}

			// if the number of producers is 0, the ET channel will close that
			// allows the Transform goroutines to terminate once they have
			// completed processing all of their data
			to.ProducerDone()
		} else if (from != nil) && (to == nil) {
			// the function is an ENDPOINT NODE of the pipeline if no data is being sent

			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
					log.Println("cluster.Load function raised error")
					supervisor.waitGroup.Done()
				}
			}()

			//aggregatedData := make([]any, 0)
			closeChan := false

			for {
				select {
				case request := <-from.GetChannel():
					{
						if request.IsInvalid() {
							closeChan = true
							break
						}

						supervisor.mutexes[fromIdx].Lock()
						supervisor.Stats.Pipes[fromIdx].Pulled++
						supervisor.mutexes[fromIdx].Unlock()

						// associates a TimeOut to the data being removed from the channel and decrements
						// the data counter for the current pipe
						from.DataPopped(request.In)

						reflect.ValueOf(supervisor.functions[functionInstance]).Call(request.Data)
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
		} else if (from != nil) && (to != nil) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("cluster.Transform function raised error")
					log.Println(r)
					to.ProducerDone()
					supervisor.waitGroup.Done()
				}
			}()

			to.AddProducer()
			closeChan := false

			for {
				select {
				case request := <-from.GetChannel():
					{
						// sometimes we are receiving bad data?
						if request.IsInvalid() {
							closeChan = true
							break
						}

						// associates a TimeOut to the data being removed from the channel and decrements
						// the data counter for the current pipe
						from.DataPopped(request.In)

						supervisor.mutexes[fromIdx].Lock()
						supervisor.Stats.Pipes[fromIdx].Pulled++
						supervisor.mutexes[fromIdx].Unlock()

						results := reflect.ValueOf(supervisor.functions[functionInstance]).Call(request.Data)
						// TODO : fix add dropping values that are bad

						supervisor.mutexes[toIdx].Lock()
						supervisor.Stats.Pipes[toIdx].Pushed++
						supervisor.mutexes[toIdx].Unlock()

						to.Push(results)
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
			to.ProducerDone()
		} else {
			// todo: clean up
			fmt.Println("not good")
		}

		// notify the wait group a process has completed ~ if all are finished we close the monitor
		supervisor.waitGroup.Done()
	}(supervisor, from, fromIdx, to, toIdx, quit)

	// a new function is provisioned
	// we should inform the wait group that the runner isn't finished until the wg is done
	supervisor.waitGroup.Add(1)
}

func (supervisor *Supervisor) RemoveProducerFrom(chanInstance int) {

	// TODO : support
	//switch segment {
	//case Transform:
	//	{
	//		if supervisor.ActiveTRoutines <= 0 {
	//			panic("attempting to quitE when no transform functions are running")
	//		}
	//		quit := supervisor.quitT[0]
	//		supervisor.quitT = supervisor.quitT[1:]
	//		quit <- true
	//	}
	//case Load:
	//	{
	//		if supervisor.ActiveLRoutines <= 0 {
	//			panic("attempting to quite when no load functions are running")
	//		}
	//		quit := supervisor.quitL[0]
	//		supervisor.quitL = supervisor.quitL[1:]
	//		quit <- true
	//	}
	//default:
	//	{
	//		panic("removing invalid cluster function type")
	//	}
	//}
}

func (supervisor *Supervisor) Deletable() bool {
	return (supervisor.State == Terminated) || (supervisor.State == Failed)
}

//func (supervisor *Supervisor) CalculateTiming() {
//
//	supervisor.Stats.Data.TotalDropped = supervisor.Stats.Data.TotalOverETChannel - supervisor.Stats.Data.TotalOverTLChannel
//	supervisor.Stats.CalculateTiming(supervisor.ETChannel.Statistics, supervisor.TLChannel.Statistics)
//}

func (supervisor *Supervisor) Print() {
	fmt.Printf("Id: %d\n", supervisor.Id)
	fmt.Printf("Function: %s\n", supervisor.Pipeline.Identifier)
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
