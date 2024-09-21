package cluster

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/channel"
)

type OneWayManagedChannel struct {
	channel *channel.ManagedChannel
}

func NewOneWayManagedChannel(c *channel.ManagedChannel) (cluster.Out, error) {

	if c == nil {
		return nil, errors.New("ManagedChannel passed to NewONeWayManagedChannel was nil")
	}

	oneWayManagedChannel := new(OneWayManagedChannel)
	oneWayManagedChannel.channel = c

	return oneWayManagedChannel, nil
}

func (c OneWayManagedChannel) Push(data any) bool {

	didPush := c.channel.Push(data)
	return didPush
}
