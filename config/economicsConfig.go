package config

// EconomicsAddresses will hold economics addresses
type EconomicsAddresses struct {
	CommunityAddress string
	BurnAddress      string
}

// RewardsSettings will hold economics rewards settings
type RewardsSettings struct {
	RewardsValue        string
	CommunityPercentage float64
	LeaderPercentage    float64
	BurnPercentage      float64
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MinGasPrice string
	MinGasLimit string
}

// RatingSettings will hold rating settings
type RatingSettings struct {
	StartRating                     int64
	MaxRating                       int64
	MinRating                       int64
	IncreaseRatingStep              int64
	DecreaseRatingStep              int64
	ProposerExtraIncreaseRatingStep int64
	ProposerExtraDecreaseRatingStep int64
}

// ConfigEconomics will hold economics config
type ConfigEconomics struct {
	EconomicsAddresses EconomicsAddresses
	RewardsSettings    RewardsSettings
	FeeSettings        FeeSettings
	RatingSettings     RatingSettings
}
