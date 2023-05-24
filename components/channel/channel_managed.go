package channel

import (
	"fmt"
	"time"
)

func NewManagedChannel(name string, threshold, growth int) *ManagedChannel {
	mc := new(ManagedChannel)

	mc.name = name
	mc.config.Threshold = threshold
	mc.config.GrowthFactor = growth
	mc.Channel = make(chan Message)

	return mc
}

func (mc *ManagedChannel) Push(data Message) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// see if we are hitting a threshold and the successive function is
	// getting overloaded with data units
	if (mc.size + 1) >= mc.config.Threshold {
		mc.state = Congested
	}
	mc.size++
	mc.lastPush = time.Now()
	mc.Channel <- data
}

func (mc *ManagedChannel) Done() {
	close(mc.Channel)
}

func (mc *ManagedChannel) AddProducer() {
	mc.wg.Add(1)
}

func (mc *ManagedChannel) ProducerDone() {

	mc.wg.Done()

	mc.wg.Wait()

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.channelFinished {
		mc.channelFinished = true
		close(mc.Channel)
	}
}

func (mc *ManagedChannel) Pull() {
	mc.size--
}

func (mc *ManagedChannel) GetState() Status {

	if mc.size == 0 {
		if time.Now().Sub(mc.lastPush).Seconds() > 3 {
			mc.state = Idle
		} else {
			mc.state = Empty
		}
	} else if mc.size > mc.config.Threshold {
		mc.state = Congested
	} else {
		mc.state = Healthy
	}

	return mc.state
}

func (mc *ManagedChannel) GetGrowthFactor() int {

	return mc.config.GrowthFactor
}

func (mc *ManagedChannel) ToString() string {
	return fmt.Sprintf("[%s][%s][Size: %d]\n", mc.name, mc.state.ToString(), mc.size)
}
