package config

// GlobalSettings will hold general economic values
type GlobalSettings struct {
	TotalSupply      string
	MinimumInflation float64
	MaximumInflation float64
}

// RewardsSettings will hold economics rewards settings
type RewardsSettings struct {
	LeaderPercentage               float64
	DeveloperPercentage            float64
	DenominationCoefficientForView string
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MaxGasLimitPerBlock  string
	GasPerDataByte       string
	DataLimitForBaseCalc string
	MinGasPrice          string
	MinGasLimit          string
}

// ValidatorSettings will hold the validator settings
type ValidatorSettings struct {
	GenesisNodePrice string
	UnBoundPeriod    string
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

//RatingValue will hold different rating options with increase and decrease steps
type RatingValue struct {
	Name  string
	Value int32
}

// EconomicsConfig will hold economics config
type EconomicsConfig struct {
	GlobalSettings    GlobalSettings
	RewardsSettings   RewardsSettings
	FeeSettings       FeeSettings
	ValidatorSettings ValidatorSettings
	RatingSettings    RatingSettings
}
