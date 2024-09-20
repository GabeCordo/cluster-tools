package interfaces

type OnCrash string

const (
	Restart   OnCrash = "Restart"
	DoNothing         = "DoNothing"
)

type OnLoad string

const (
	CompleteAndPush OnLoad = "CompleteAndPush"
	WaitAndPush            = "WaitAndPush"
)

type EtlMode string

const (
	Batch  EtlMode = "Batch"
	Stream         = "Stream"
)

type Config struct {
	Identifier                  string  `json:"identifier"`
	OnLoad                      OnLoad  `json:"on-load"`
	OnCrash                     OnCrash `json:"on-crash"`
	StartWithNTransformClusters int     `json:"start-with-n-t-channels"`
	StartWithNLoadClusters      int     `json:"start-with-n-l-channels"`
	ETChannelThreshold          int     `json:"et-channel-threshold"`
	ETChannelGrowthFactor       float64 `json:"et-channel-growth-factor"`
	TLChannelThreshold          int     `json:"tl-channel-threshold"`
	TLChannelGrowthFactor       float64 `json:"tl-channel-growth-factor"`
}
