package channel

type BadManagedChannelType struct {
	description string
}

func (bmce BadManagedChannelType) Error() string {
	return bmce.description
}
