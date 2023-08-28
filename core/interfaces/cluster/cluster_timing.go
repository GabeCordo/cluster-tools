package cluster

func (dataTiming DataTiming) Valid() bool {
	return !dataTiming.ETIn.IsZero() && !dataTiming.ETOut.IsZero() && !dataTiming.TLIn.IsZero() && !dataTiming.TLOut.IsZero()
}
