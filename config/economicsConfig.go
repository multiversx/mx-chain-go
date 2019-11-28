package config

// EconomicsAddresses will hold economics addresses
type EconomicsAddresses struct {
	CommunityAddress string
	BurnAddress      string
}

// RewardsSettings will hold economics rewards settings
type RewardsSettings struct {
	RewardsValue                   string
	CommunityPercentage            float64
	LeaderPercentage               float64
	BurnPercentage                 float64
	DenominationCoefficientForView string
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MaxGasLimitPerBlock string
	MinGasPrice         string
	MinGasLimit         string
}

// ValidatorSettings will hold the validator settings
type ValidatorSettings struct {
	StakeValue    string
	UnBoundPeriod string
}

// ConfigEconomics will hold economics config
type ConfigEconomics struct {
	EconomicsAddresses EconomicsAddresses
	RewardsSettings    RewardsSettings
	FeeSettings        FeeSettings
	ValidatorSettings  ValidatorSettings
}
