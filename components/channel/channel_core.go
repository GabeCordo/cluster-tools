package channel

func NewManagedChannel(threshold, growth int) *ManagedChannel {
	mc := new(ManagedChannel)

	mc.Config.Threshold = threshold
	mc.Config.GrowthFactor = growth
	mc.Channel = make(chan Message)

	return mc
}

func (mc *ManagedChannel) Push(data Message) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// see if we are hitting a threshold and the successive function is
	// getting overloaded with data units
	if (mc.Config.Size + 1) >= mc.Config.Threshold {
		mc.State = Congested
	}
	mc.Config.Size++
	mc.Channel <- data

	mc.wg.Add(1)
}

func (mc *ManagedChannel) Pull() Message {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	holder := <-mc.Channel // block until data is pulled
	mc.Config.Size--
	if mc.Config.Size == 0 {
		mc.State = Empty
	}

	mc.wg.Done()
	return holder
}

func (mc *ManagedChannel) IfEmptyProceed() {
	// if the channel is empty, proceed
	mc.wg.Wait()
}
