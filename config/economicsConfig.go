package config

// GlobalSettings will hold general economic values
type GlobalSettings struct {
	GenesisTotalSupply          string
	MinimumInflation            float64
	YearSettings                []*YearSetting
	TailInflation               TailInflationSettings
	Denomination                int
	GenesisMintingSenderAddress string
}

// TailInflationSettings will hold the tail inflation settings
type TailInflationSettings struct {
	ActivationEpoch        uint32
	MaximumYearlyInflation float64
	DecayPercentage        float64
	MinimumInflation       float64
}

// YearSetting will hold the maximum inflation rate for year
type YearSetting struct {
	Year             uint32
	MaximumInflation float64
}

// RewardsSettings holds the economics rewards config changes by epoch
type RewardsSettings struct {
	RewardsConfigByEpoch []EpochRewardSettings
	TailInflation        TailInflationSettings
}

// EpochRewardSettings holds the economics rewards settings for a specific epoch
type EpochRewardSettings struct {
	LeaderPercentage                 float64
	DeveloperPercentage              float64
	ProtocolSustainabilityPercentage float64
	ProtocolSustainabilityAddress    string
	EcosystemGrowthPercentage        float64
	EcosystemGrowthAddress           string
	GrowthDividendPercentage         float64
	GrowthDividendAddress            string
	TopUpGradientPoint               string
	TopUpFactor                      float64
	EpochEnable                      uint32
}

// GasLimitSetting will hold gas limit setting for a specific epoch
type GasLimitSetting struct {
	EnableEpoch                 uint32
	MaxGasLimitPerBlock         string
	MaxGasLimitPerMiniBlock     string
	MaxGasLimitPerMetaBlock     string
	MaxGasLimitPerMetaMiniBlock string
	MaxGasLimitPerTx            string
	MinGasLimit                 string
	ExtraGasLimitGuardedTx      string
	MaxGasHigherFactorAccepted  string
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	GasLimitSettings       []GasLimitSetting
	GasPerDataByte         string
	MinGasPrice            string
	GasPriceModifier       float64
	MaxGasPriceSetGuardian string
}

// EconomicsConfig will hold economics config
type EconomicsConfig struct {
	GlobalSettings  GlobalSettings
	RewardsSettings RewardsSettings
	FeeSettings     FeeSettings
}
