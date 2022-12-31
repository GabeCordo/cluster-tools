package cluster

func NewStatistics() *Statistics {
	stats := new(Statistics)

	stats.NumProvisionedTransformRoutes = 0
	stats.NumProvisionedLoadRoutines = 0
	stats.NumTlThresholdBreaches = 0
	stats.NumEtThresholdBreaches = 0

	return stats
}
