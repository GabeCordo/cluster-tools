package scheduler

func (filter Filter) UseIdentifier() bool {
	return filter.Identifier != ""
}

func (filter Filter) UseModule() bool {
	return filter.Interval.Empty() && (filter.Module != "") && (filter.Cluster == "")
}

func (filter Filter) UseCluster() bool {
	return filter.Interval.Empty() && (filter.Module != "") && (filter.Cluster != "")
}

func (filter Filter) UseInterval() bool {

	return !filter.Interval.Empty() && (filter.Module != "") && (filter.Cluster != "")
}
