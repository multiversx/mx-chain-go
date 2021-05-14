package config

// GlobalSettings will hold general economic values
type GlobalSettings struct {
	GenesisTotalSupply string
	MinimumInflation   float64
	YearSettings       []*YearSetting
	Denomination       int
}

// YearSetting will hold the maximum inflation rate for year
type YearSetting struct {
	Year             uint32
	MaximumInflation float64
}

// RewardsSettings holds the economics rewards config changes by epoch
type RewardsSettings struct {
	RewardsConfigByEpoch []EpochRewardSettings
}

// RewardsConfig holds the economics rewards settings for a specific epoch
type EpochRewardSettings struct {
	LeaderPercentage                 float64
	DeveloperPercentage              float64
	ProtocolSustainabilityPercentage float64
	ProtocolSustainabilityAddress    string
	TopUpGradientPoint               string
	TopUpFactor                      float64
	EpochEnable                      uint32
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MaxGasLimitPerBlock     string
	MaxGasLimitPerMetaBlock string
	GasPerDataByte          string
	MinGasPrice             string
	MinGasLimit             string
	GasPriceModifier        float64
}

// EconomicsConfig will hold economics config
type EconomicsConfig struct {
	GlobalSettings  GlobalSettings
	RewardsSettings RewardsSettings
	FeeSettings     FeeSettings
}
