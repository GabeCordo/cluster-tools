package cluster

import (
	"fmt"
)

var DefaultConfig Config = Config{
	Identifier:                  "",
	OnCrash:                     DoNothing,
	OnLoad:                      CompleteAndPush,
	StartWithNTransformClusters: 1,
	StartWithNLoadClusters:      1,
	ETChannelThreshold:          2,
	ETChannelGrowthFactor:       2,
	TLChannelThreshold:          2,
	TLChannelGrowthFactor:       2,
}

func NewConfig(identifier string, etChannelThreshold, etChannelGrowthFactor, tlChannelThreshold, tlChannelGrowthFactor int, mode OnCrash) *Config {
	config := new(Config)

	config.Identifier = identifier
	config.ETChannelThreshold = etChannelThreshold
	config.ETChannelGrowthFactor = etChannelGrowthFactor
	config.TLChannelThreshold = tlChannelThreshold
	config.TLChannelGrowthFactor = tlChannelGrowthFactor
	config.OnCrash = mode
	config.OnLoad = CompleteAndPush

	return config
}

func (config Config) Valid() bool {
	return !((config.StartWithNLoadClusters <= 0) || (config.StartWithNTransformClusters <= 0) ||
		(config.TLChannelThreshold < 1) || (config.ETChannelThreshold < 1) ||
		(config.TLChannelGrowthFactor <= 1) || (config.ETChannelGrowthFactor <= 1))
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
