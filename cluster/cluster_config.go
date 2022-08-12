package cluster

func NewConfig(identifier string, etChannelThreshold, etChannelGrowthFactor, tlChannelThreshold, tlChannelGrowthFactor int) *Config {
	config := new(Config)

	config.Identifier = identifier
	config.etChannelGrowthFactor = etChannelGrowthFactor
	config.tlChannelThreshold = tlChannelThreshold
	config.tlChannelGrowthFactor = tlChannelGrowthFactor

	return config
}
