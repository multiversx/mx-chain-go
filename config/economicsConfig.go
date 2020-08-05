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

// RewardsSettings will hold economics rewards settings
type RewardsSettings struct {
	LeaderPercentage                 float64
	DeveloperPercentage              float64
	ProtocolSustainabilityPercentage float64
	ProtocolSustainabilityAddress    string
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MaxGasLimitPerBlock     string
	MaxGasLimitPerMetaBlock string
	GasPerDataByte          string
	DataLimitForBaseCalc    string
	MinGasPrice             string
	MinGasLimit             string
}

// EconomicsConfig will hold economics config
type EconomicsConfig struct {
	GlobalSettings  GlobalSettings
	RewardsSettings RewardsSettings
	FeeSettings     FeeSettings
}
