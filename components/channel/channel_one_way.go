package channel

func NewOneWayManagedChannel(channel *ManagedChannel) (*OneWayManagedChannel, error) {

	if channel == nil {
		return nil, BadManagedChannelType{description: "ManagedChannel passed to NewONeWayManagedChannel was nil"}
	}

	oneWayManagedChannel := new(OneWayManagedChannel)
	oneWayManagedChannel.channel = channel

	return oneWayManagedChannel, nil
}

func (owmc *OneWayManagedChannel) Push(data any) {

	owmc.channel.channel <- data
	owmc.channel.Size++
}
