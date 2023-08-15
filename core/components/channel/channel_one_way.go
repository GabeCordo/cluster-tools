package channel

import (
	"github.com/GabeCordo/etl-light/components/channel"
	"math/rand"
)

func NewOneWayManagedChannel(channel *ManagedChannel) (channel.OneWay, error) {

	if channel == nil {
		return nil, BadManagedChannelType{description: "ManagedChannel passed to NewONeWayManagedChannel was nil"}
	}

	oneWayManagedChannel := new(OneWayManagedChannel)
	oneWayManagedChannel.channel = channel

	return oneWayManagedChannel, nil
}

func (owmc OneWayManagedChannel) Push(data any) {

	owmc.channel.Push(DataWrapper{Id: rand.Uint64(), Data: data})
}
