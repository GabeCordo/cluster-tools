package channel

import (
	"fmt"
	"time"
)

func NewManagedChannel(name string, threshold, growth int) *ManagedChannel {
	mc := new(ManagedChannel)

	mc.Name = name
	mc.Config.Threshold = threshold
	mc.Config.GrowthFactor = growth
	mc.channel = make(chan any)

	return mc
}

func (mc *ManagedChannel) GetChannel() chan any {
	return mc.channel
}

func (mc *ManagedChannel) Push(data any) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// see if we are hitting a threshold and the successive function is
	// getting overloaded with data units
	if (mc.Size + 1) >= mc.Config.Threshold {
		mc.State = Congested
	}
	mc.Size++
	mc.LastPush = time.Now()
	mc.channel <- data
}

func (mc *ManagedChannel) Done() {
	close(mc.channel)
}

func (mc *ManagedChannel) AddProducer() {
	mc.wg.Add(1)
}

func (mc *ManagedChannel) ProducerDone() {

	mc.wg.Done()

	mc.wg.Wait()

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.ChannelFinished {
		mc.ChannelFinished = true
		close(mc.channel)
	}
}

func (mc *ManagedChannel) Pull() {
	mc.Size--
}

func (mc *ManagedChannel) GetState() Status {

	if mc.Size == 0 {
		if time.Now().Sub(mc.LastPush).Seconds() > 3 {
			mc.State = Idle
		} else {
			mc.State = Empty
		}
	} else if mc.Size > mc.Config.Threshold {
		mc.State = Congested
	} else {
		mc.State = Healthy
	}

	return mc.State
}

func (mc *ManagedChannel) GetGrowthFactor() int {

	return mc.Config.GrowthFactor
}

func (mc *ManagedChannel) ToString() string {
	return fmt.Sprintf("[%s][%s][Size: %d]\n", mc.Name, mc.State.ToString(), mc.Size)
}
