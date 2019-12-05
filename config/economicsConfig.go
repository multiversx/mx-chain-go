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

// RatingSettings will hold rating settings
type RatingSettings struct {
	StartRating                 uint32
	MaxRating                   uint32
	MinRating                   uint32
	ProposerIncreaseRatingStep  uint32
	ProposerDecreaseRatingStep  uint32
	ValidatorIncreaseRatingStep uint32
	ValidatorDecreaseRatingStep uint32
}

//RatingValue will hold different rating options with increase and decresea steps
type RatingValue struct {
	Name  string
	Value int32
}

// ConfigEconomics will hold economics config
type ConfigEconomics struct {
	EconomicsAddresses EconomicsAddresses
	RewardsSettings    RewardsSettings
	FeeSettings        FeeSettings
	ValidatorSettings  ValidatorSettings
	RatingSettings     RatingSettings
}
