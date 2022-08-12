package cluster

func NewStatistics(numProvisionedTransformRoutes, numProvisionedLoadRoutines int) *Statistics {
	stats := new(Statistics)

	stats.NumProvisionedTransformRoutes = numProvisionedTransformRoutes
	stats.NumProvisionedLoadRoutines = numProvisionedLoadRoutines

	return stats
}
