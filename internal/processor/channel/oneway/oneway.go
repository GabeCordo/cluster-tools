package oneway

import (
	"errors"
	"github.com/GabeCordo/cluster-tools/cluster"
	"github.com/GabeCordo/cluster-tools/internal/processor/channel/duplex"
)

type ManagedChannel struct {
	channel *duplex.ManagedChannel
}

func NewOneWayManagedChannel(c *duplex.ManagedChannel) (cluster.Out, error) {

	if c == nil {
		return nil, errors.New("ManagedChannel passed to NewONeWayManagedChannel was nil")
	}

	oneWayManagedChannel := new(ManagedChannel)
	oneWayManagedChannel.channel = c

	return oneWayManagedChannel, nil
}

func (c ManagedChannel) Push(data any) bool {

	didPush := c.channel.Push(data)
	return didPush
}
