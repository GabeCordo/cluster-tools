package cluster

import "fmt"

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

func (config Config) Print() {
	fmt.Printf("Identifier:\t%s\n", config.Identifier)
	fmt.Printf("StartWithNTransform:\t%d\n", config.StartWithNTransformClusters)
	fmt.Printf("StartWithNLoad:\t%d\n", config.StartWithNLoadClusters)
	fmt.Printf("ETChannelThreshold:\t%d\n", config.ETChannelThreshold)
	fmt.Printf("ETChannelGrowthFactor:\t%d\n", config.ETChannelGrowthFactor)
	fmt.Printf("TLChannelThreshold:\t%d\n", config.TLChannelThreshold)
	fmt.Printf("TLChannelGrowthFactor:\t%d\n", config.TLChannelGrowthFactor)
}
