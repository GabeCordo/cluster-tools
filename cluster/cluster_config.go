package cluster

func NewConfig(identifier string, etChannelThreshold, etChannelGrowthFactor, tlChannelThreshold, tlChannelGrowthFactor int, mode OnCrash) *Config {
	config := new(Config)

	config.Identifier = identifier
	config.ETChannelThreshold = etChannelThreshold
	config.ETChannelGrowthFactor = etChannelGrowthFactor
	config.TLChannelThreshold = tlChannelThreshold
	config.TLChannelGrowthFactor = tlChannelGrowthFactor
	config.Mode = mode

	return config
}
