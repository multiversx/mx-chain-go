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
	MinGasPrice string
	MinGasLimit string
}

// ConfigEconomics will hold economics config
type ConfigEconomics struct {
	EconomicsAddresses EconomicsAddresses
	RewardsSettings    RewardsSettings
	FeeSettings        FeeSettings
}
