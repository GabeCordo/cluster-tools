package pipeline

import (
	"fmt"
	pipeline_cfg "github.com/GabeCordo/cluster-tools/internal/core/database/pipeline"
	"github.com/GabeCordo/cluster-tools/internal/core/database/statistic"
	"github.com/GabeCordo/cluster-tools/internal/processor/channel/duplex"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
)

const (
	DefaultNumberOfClusters       = 1
	DefaultMonitorRefreshDuration = 100
	DefaultChannelThreshold       = 10
	DefaultChannelGrowthFactor    = 2
)

type Response struct {
	Stats      *statistic.Statistics `json:"stats"`
	LapsedTime time.Duration         `json:"lapsed-time"`
	DidItCrash bool                  `json:"crashed"`
}

func NewResponse(statistics *statistic.Statistics, lapsedTime time.Duration, crashed bool) *Response {
	response := new(Response)

	response.Stats = statistics
	response.LapsedTime = lapsedTime
	response.DidItCrash = crashed

	return response
}

type Status string

const (
	UnTouched    Status = "untouched"
	Active              = "active"
	Provisioning        = "provisioning"
	Failed              = "failed"
	Stopping            = "stopping"
	Terminated          = "terminated"
	Unknown             = "-"
)

type Event uint8

const (
	Startup Event = iota
	StartProvision
	EndProvision
	Error
	Suspend
	TearedDown
	StartReport
	EndReport
)

const MaximumRoutinesPerSupervisor = 2000

type Instance struct {
	Id uint64 `json:"id"`

	State     Status    `json:"status"`
	StartTime time.Time `json:"quitE-time"`

	Pipeline  *Pipeline
	metadata  map[string]string
	functions []any // initialized in new

	loadWaitGroup sync.WaitGroup
	waitGroup     sync.WaitGroup
	threadMutex   sync.Mutex
	mutex         sync.RWMutex
}

func NewInstance(config *pipeline_cfg.Pipeline, functions []any, metadata map[string]string) *Instance {
	supervisor := new(Instance)

	/**
	 * Note: we may wish to dynamically modify the threshold and growth-factor rates
	 *       used by the managed channels to vary how provisioning of new transform and
	 *       load goroutines are created. This allows us to create an autonomous system
	 *       that "self improves" if the output of the monitor is looped back
	 */

	supervisor.State = UnTouched

	supervisor.functions = functions
	supervisor.Pipeline = New(config, functions)
	supervisor.metadata = metadata

	return supervisor
}

type Summary struct {
	Namespace  string
	Pipeline   string
	Supervisor uint64
	Statistics *statistic.Statistics
}

func (supervisor *Instance) Event(event Event) bool {
	supervisor.mutex.Lock()
	defer supervisor.mutex.Unlock()

	if supervisor.State == UnTouched {
		if event == Startup {
			supervisor.State = Active
		} else if (event == Suspend) || (event == TearedDown) {
			supervisor.State = Stopping
		} else {
			return false
		}
	} else if supervisor.State == Active {
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
			supervisor.State = Active
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

func (supervisor *Instance) IsAlive() bool {

	supervisor.mutex.RLock()
	defer supervisor.mutex.RUnlock()

	return (supervisor.State != Failed) && (supervisor.State != Terminated)
}

func (supervisor *Instance) Start() (response *Response) {
	supervisor.Event(Startup)

	defer supervisor.Event(TearedDown)

	defer func() {
		// has the user defined function crashed during runtime?
		if r := recover(); r != nil {
			// yes => return a response that identifies that the cluster crashed
			response = NewResponse(
				supervisor.Pipeline.Stats,
				time.Now().Sub(supervisor.StartTime),
				true,
			)
		}
	}()

	supervisor.StartTime = time.Now()

	if supervisor.Pipeline.OnStartup != nil {
		// TODO : add safety check here
		f := reflect.ValueOf(supervisor.Pipeline.OnStartup.Value)
		f.Call([]reflect.Value{})
	}

	//// add all the metadata passed to the pipeline to the local environment

	// TODO : possibly enhance security?
	for key, value := range supervisor.metadata {
		os.Setenv(key, value)
	}

	//// start creating the default frontend goroutines

	for _, function := range supervisor.Pipeline.Functions {
		for j := 0; j < function.Config.StartWith; j++ {
			supervisor.Provision(function)
			function.Stats.Active++
			function.Stats.Provisions++
		}
	}

	//// end creating the default frontend goroutines

	// every N seconds we should check if the ETChannel or TLChannel is congested
	// and requires us to provision additional nodes
	go supervisor.Runtime()

	supervisor.waitGroup.Wait() // wait for the Extract-Transform-Load (ETL) Cycle to Complete

	// calculate the timings produced by data being fed across each of the channels
	// TODO: support
	//supervisor.CalculateTiming()

	response = NewResponse(
		supervisor.Pipeline.Stats,
		time.Now().Sub(supervisor.StartTime),
		false,
	)

	//// cleanup environment variables that were dynamically set

	for key, _ := range supervisor.metadata {
		os.Unsetenv(key)
	}

	return response
}

func (supervisor *Instance) Teardown() {

	// TODO : add a guard in case this value is not a function
	if supervisor.Pipeline.OnStartup != nil {
		f := reflect.ValueOf(supervisor.Pipeline.OnTeardown.Value)
		f.Call([]reflect.Value{})
	}

	supervisor.Event(Suspend)
}

func (supervisor *Instance) Runtime() {
	for {
		if supervisor.State == Terminated {
			break
		}

		for _, channel := range supervisor.Pipeline.Channels {

			channelState := channel.Value.GetState()

			if (supervisor.State == Stopping) && channel.Value.Accepting() {
				channel.Value.StopPushes()
			}

			if channelState == duplex.Congested {

				channel.Stats.Breaches++

				for _, f := range channel.Receiver {
					n := channel.Config.GrowthFactor
					for n > 0 {
						f.Stats.Provisions++
						f.Stats.Active++
						supervisor.Provision(f)
						n--
					}
				}
			} else if (channelState == duplex.Underutilized) || (channelState == duplex.Idle) {

				for _, f := range channel.Receiver {

					n := channel.Config.GrowthFactor
					for n > 0 {
						// never remove all transform nodes otherwise we risk the
						// ET channel having no consumers
						if f.Stats.Active <= 1 {
							break
						}
						f.Stats.Active--
						supervisor.Remove(f)
						n--
					}
				}
			}
		}

		// check if the channel is congested after DefaultMonitorRefreshDuration seconds
		time.Sleep(DefaultMonitorRefreshDuration * time.Millisecond)
	}
}

func (supervisor *Instance) ExtractWrapper(function any, channel *duplex.ManagedChannel) <-chan struct{} {
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
				channel := reflect.MakeChan(channelType, 0)
				arguments = append(arguments, channel)
			}
		}

		go reflect.ValueOf(function).Call(arguments)

		if reflect.TypeOf(function).NumIn() > 0 {

			for {
				value, ok := arguments[0].Recv()
				if ok {
					channel.Push([]reflect.Value{value})
				} else {
					break
				}
			}
		}
	}()
	return done
}

func (supervisor *Instance) ExtractShutdownWrapper() <-chan struct{} {
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

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// Call
// a wrapper function that checks the returned values from a function call for
// error values. if the function returns an error that is non-nil we will set
// the returned boolean flag to true indicating something may have gone wrong
// inside the function call.
func (supervisor *Instance) Call(function *Function, in []reflect.Value) ([]reflect.Value, bool) {

	drop := false
	results := reflect.ValueOf(function.Value).Call(in)

	// TODO : the number of results returned by a function can be pre-computed
	numResults := len(results)

	// a common pattern in golang is to return (value, error) where error can be used
	// to identify that the function was not able to run to completion.
	//
	// the pipeline supports a similar method of identifying that the function failed
	// to run successfully to completion. when the function returns an error value
	// (that must be the last value returned) the pipeline will check whether the
	// error is nil or not to determine whether the resultant data should be used.
	//
	// error is nil -> keep sending the resultant data along the pipeline
	// error -> do NOT pass the data along the pipeline (aka. drop the data)
	if numResults > 0 {

		lastResult := results[numResults-1]

		isError := lastResult.Type().Implements(errorInterface)

		if isError {
			isNil := lastResult.IsNil()

			if !isNil {
				drop = true
			}

			// never pass an error along the pipeline
			results = results[:numResults-1]
		}
	}

	return results, drop
}

func (supervisor *Instance) Provision(function *Function) {
	supervisor.Event(StartProvision)
	defer supervisor.Event(EndProvision)

	// the runner must provision new threads one at a time.
	// note: avoid the possibility of >1 thread modifying the wait-group at a time
	supervisor.threadMutex.Lock()
	defer supervisor.threadMutex.Unlock()

	go func(supervisor *Instance, function *Function) {

		if (function.From == nil) && (function.To != nil) {
			// the function is a STARTING NODE of the Pipeline if no data is being received

			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
					log.Println("cluster.Extract function raised error")
					function.To.Value.ProducerDone()
					supervisor.waitGroup.Done()
				}
			}()

			function.To.Value.AddProducer()

			// TODO : reword
			// if the producer function has output, then don't worry about spawning a wrapper,
			// allow the function to return normally and send the data along the pipe

			if reflect.TypeOf(function.Value).NumOut() > 0 {

				output := reflect.ValueOf(function.Value).Call([]reflect.Value{})
				function.To.Value.Push(output)

			} else {

				select {
				case <-supervisor.ExtractWrapper(function.Value, function.To.Value):
					break
				case <-supervisor.ExtractShutdownWrapper():
					fmt.Println("shutdown caused extract to finish early")
					break
				}
			}

			// if the number of producers is 0, the ET channel will close that
			// allows the Transform goroutines to terminate once they have
			// completed processing all of their data
			function.To.Value.ProducerDone()
		} else if (function.From != nil) && (function.To == nil) {
			// the function is an ENDPOINT NODE of the Pipeline if no data is being sent

			quit := make(chan bool)
			function.Mutex.Lock()
			function.Quit = append(function.Quit, quit)
			function.Mutex.Unlock()

			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
					log.Println("cluster.Load function raised error")
					supervisor.waitGroup.Done()
				}
			}()

			// todo ?
			if function.Config.WaitBefore {

			}

			var queuedRequests reflect.Value
			if function.Config.WaitBefore {
				in := reflect.TypeOf(function.Value).In(0)
				queuedRequests = reflect.MakeSlice(in, 0, 0)
			}
			closeChan := false

			for {
				select {
				case request := <-function.From.Value.GetChannel():
					{
						if request.IsInvalid() {
							closeChan = true
							break
						}

						function.From.Mutex.Lock()
						function.From.Stats.Pulled++
						function.From.Mutex.Unlock()

						// associates a TimeOut to the data being removed from the channel and decrements
						// the data counter for the current pipe
						function.From.Value.DataPopped(request.In)

						// what: the developer has an option to wait for all the data to be received by a
						// channel before processing that data.
						//
						// why: we may need a bulk set of data before being able to do anything
						if function.Config.WaitBefore {
							queuedRequests = reflect.Append(queuedRequests, request.Data[0])
						} else {
							reflect.ValueOf(function.Value).Call(request.Data)
						}
					}
				case <-quit:
					{
						closeChan = true
					}
				}

				if closeChan {

					// if we were waiting for the channel to close before transforming the data,
					// call the function now that the channel is closed
					if function.Config.WaitBefore {
						reflect.ValueOf(function.Value).Call([]reflect.Value{queuedRequests})
					}

					break
				}
			}
		} else if (function.From != nil) && (function.To != nil) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("cluster.Transform function raised error")
					log.Println(r)
					function.To.Value.ProducerDone()
					supervisor.waitGroup.Done()
				}
			}()

			quit := make(chan bool)
			function.Mutex.Lock()
			function.Quit = append(function.Quit, quit)
			function.Mutex.Unlock()

			function.To.Value.AddProducer()

			var queuedRequests reflect.Value
			if function.Config.WaitBefore {
				in := reflect.TypeOf(function.Value).In(0)
				queuedRequests = reflect.MakeSlice(in, 0, 0)
			}
			closeChan := false

			for {
				select {
				case request := <-function.From.Value.GetChannel():
					{
						// sometimes we are receiving bad data?
						if request.IsInvalid() {
							closeChan = true
							break
						}

						// associates a TimeOut to the data being removed from the channel and decrements
						// the data counter for the current pipe
						function.From.Value.DataPopped(request.In)

						function.From.Mutex.Lock()
						function.From.Stats.Pulled++
						function.From.Mutex.Unlock()

						// what: the developer has an option to wait for all the data to be received by a
						// channel before processing that data.
						//
						// why: we may need a bulk set of data before being able to do anything
						if function.Config.WaitBefore {
							queuedRequests = reflect.Append(queuedRequests, request.Data[0])
						} else {
							results, drop := supervisor.Call(function, request.Data)

							function.To.Mutex.Lock()
							if !drop {

								// PROPOSAL 6.
								// ~let there be two functions f1 and f2 that are transformers along the pipeline.
								// ~let f1 return a slice of type A s.t. []A is the passed along value
								// ~let f2 accept a value of type A s.t. we expect f1 to send us A
								//
								// if (f1 sends to f2) and (f1 returns []A while f2 accepts A) then:
								// the pipeline shall break apart the []A returned by f1
								// (and) push each element A from the slice to the successive function f2
								if len(results) > 0 && results[0].Kind() == reflect.Slice {

									// note: we should always have some receiver to pull data from the channel but
									//		 this is a safety guard if something goes wrong
									if len(function.To.Receiver) > 0 {
										nextFunction := function.To.Receiver[0]
										nextFunctionReflection := reflect.TypeOf(nextFunction.Value)

										// note: the value should accept some value or the pipeline we've provisioned
										// 		 is invalid and should have been rejected prior to this step
										if (nextFunctionReflection.NumIn() > 0) && (nextFunctionReflection.In(0).Kind() != reflect.Slice) {

											// f1 returns []A and f2 accepts A has been validated upto this point
											// we should break apart []A and send the data 1-by-1 to f2
											for i := 0; i < results[0].Len(); i++ {
												function.To.Stats.Pushed++
												function.To.Value.Push([]reflect.Value{results[0].Index(i)})
											}
										} else {
											// default functionality
											function.To.Stats.Pushed++
											function.To.Value.Push(results)
										}
									} else {
										// default functionality
										function.To.Stats.Pushed++
										function.To.Value.Push(results)
									}

								} else {
									// default functionality
									function.To.Stats.Pushed++
									function.To.Value.Push(results)
								}

							} else {
								function.From.Stats.Dropped++
							}
							function.To.Mutex.Unlock()

						}
					}
				case <-quit:
					{
						closeChan = true
					}
				}

				if closeChan {

					// if we were waiting for the channel to close before transforming the data,
					// call the function now that the channel is closed
					if function.Config.WaitBefore {
						results, drop := supervisor.Call(function, []reflect.Value{queuedRequests})

						function.To.Mutex.Lock()
						if !drop {
							function.To.Stats.Pushed++
							function.To.Value.Push(results)
						} else {
							function.From.Stats.Dropped++
						}
						function.To.Mutex.Unlock()
					}
					break
				}
			}

			// if the number of producers is 0, the TL channel will close that
			// allows the Load goroutines to terminate once they have
			// completed processing all of their data
			function.To.Value.ProducerDone()
		} else {
			// todo: clean up
			fmt.Println("not good")
		}

		// notify the wait group a process has completed ~ if all are finished we close the monitor
		supervisor.waitGroup.Done()
	}(supervisor, function)

	// a new function is provisioned
	// we should inform the wait group that the runner isn't finished until the wg is done
	supervisor.waitGroup.Add(1)
}

func (supervisor *Instance) Remove(function *Function) {

	if function.Stats.Active <= 0 {
		panic("attempting to quit when no functions are running")
	}

	quit := function.Quit[0]
	function.Quit = function.Quit[1:]
	quit <- true
}

func (supervisor *Instance) Deletable() bool {
	return (supervisor.State == Terminated) || (supervisor.State == Failed)
}

//func (supervisor *Instance) CalculateTiming() {
//
//	supervisor.Stats.Data.TotalDropped = supervisor.Stats.Data.TotalOverETChannel - supervisor.Stats.Data.TotalOverTLChannel
//	supervisor.Stats.CalculateTiming(supervisor.ETChannel.Statistics, supervisor.TLChannel.Statistics)
//}

func (supervisor *Instance) Print() {
	fmt.Printf("Id: %d\n", supervisor.Id)
	// TODO : fix
	//fmt.Printf("Function: %s\n", supervisor.Pipeline)
}

func (status Status) ToString() string {
	switch status {
	case UnTouched:
		return "UnTouched"
	case Active:
		return "Active"
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
