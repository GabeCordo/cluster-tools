package supervisor

func (filter Filter) UseId() bool {
	return filter.Id != 0
}

func (filter Filter) UseModule() bool {
	return (filter.Module != "") && (filter.Cluster == "")
}

func (filter Filter) UseCluster() bool {
	return (filter.Module != "") && (filter.Cluster != "")
}
