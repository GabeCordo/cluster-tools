package interfaces

import "time"

type DataTimer struct {
	In  time.Time
	Out time.Time
}

func (dataTiming DataTiming) Valid() bool {
	return !dataTiming.ETIn.IsZero() && !dataTiming.ETOut.IsZero() && !dataTiming.TLIn.IsZero() && !dataTiming.TLOut.IsZero()
}
