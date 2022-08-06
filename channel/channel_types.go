package channel

type Status int

const (
	Empty     Status = 0
	Healthy          = 1
	Congested        = 2
)

type Message any

type OutputChannel chan<- Message

type InputChannel <-chan Message

type ManagedChannelConfig struct {
	Threshold    int
	GrowthFactor int
	Size         int
}

type ManagedChannel struct {
	State   Status
	Config  ManagedChannelConfig
	Channel chan Message
}
