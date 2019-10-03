package config

// EconomicsAddresses will hold economics addresses
type EconomicsAddresses struct {
	CommunityAddress string
	BurnAddress      string
}

// RewardsSettings will hold economics rewards settings
type RewardsSettings struct {
	RewardsValue        uint64
	CommunityPercentage float64
	LeaderPercentage    float64
	BurnPercentage      float64
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MinGasPrice      uint64
	MinGasLimitForTx uint64
	MinTxFee         uint64
}

// ConfigEconomics will hold economics config
type ConfigEconomics struct {
	EconomicsAddresses EconomicsAddresses
	RewardsSettings    RewardsSettings
	FeeSettings        FeeSettings
}
