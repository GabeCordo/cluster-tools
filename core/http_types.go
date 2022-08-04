package core

// Frontend Thread

type Http struct {
	Interrupt chan<- InterruptEvent // Upon completion or failure an interrupt can be raised

	C1 chan<- DatabaseRequest  // Core is sending core to the Database
	C2 <-chan DatabaseResponse // Core is receiving responses from the Database

	C5 chan<- SupervisorRequest  // Core is sending core to the Database
	C6 <-chan SupervisorResponse // Core is receiving responses from the Database
}

func NewHttp(channels ...interface{}) (*Http, bool) {
	core := new(Http)
	var ok bool

	core.Interrupt, ok = (channels[0]).(chan InterruptEvent)
	if !ok {
		return nil, ok
	}
	core.C1, ok = (channels[1]).(chan DatabaseRequest)
	if !ok {
		return nil, ok
	}
	core.C2, ok = (channels[2]).(chan DatabaseResponse)
	if !ok {
		return nil, ok
	}
	core.C5, ok = (channels[3]).(chan SupervisorRequest)
	if !ok {
		return nil, ok
	}
	core.C6, ok = (channels[4]).(chan SupervisorResponse)
	if !ok {
		return nil, ok
	}

	return core, ok
}
