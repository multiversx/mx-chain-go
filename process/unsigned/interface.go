package unsigned

// RewardsHandler will return information about rewards
type RewardsHandler interface {
	RewardsValue() uint64
	CommunityPercentage() float64
	LeaderPercentage() float64
	BurnPercentage() float64
}

// FeeHandler will return information about fees
type FeeHandler interface {
	MinGasPrice() uint64
	MinGasLimitForTx() uint64
	MinTxFee() uint64
}

// EconomicsAddressesHandler will return information about economics addresses
type EconomicsAddressesHandler interface {
	CommunityAddress() string
	BurnAddress() string
}
