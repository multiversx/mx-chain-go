package config

// RatingsConfig will hold the configuration data needed for the ratings
type RatingsConfig struct {
	General     General
	ShardChain  ShardChain
	MetaChain   MetaChain
	PeerHonesty PeerHonestyConfig
}

// General will hold ratings settings both for metachain and shardChain
type General struct {
	StartRating           uint32
	MaxRating             uint32
	MinRating             uint32
	SignedBlocksThreshold float32
	SelectionChances      []*SelectionChance
}

// ShardChain will hold RatingSteps for the Shard
type ShardChain struct {
	RatingSteps
}

// MetaChain will hold RatingSteps for the Meta
type MetaChain struct {
	RatingSteps
}

//RatingValue will hold different rating options with increase and decrease steps
type RatingValue struct {
	Name  string
	Value int32
}

// SelectionChance will hold the percentage modifier for up to the specified threshold
type SelectionChance struct {
	MaxThreshold  uint32
	ChancePercent uint32
}

// RatingSteps holds the necessary increases and decreases of the rating steps
type RatingSteps struct {
	HoursToMaxRatingFromStartRating uint32
	ProposerValidatorImportance     float32
	ProposerDecreaseFactor          float32
	ValidatorDecreaseFactor         float32
	ConsecutiveMissedBlocksPenalty  float32
}

// PeerHonestyConfig holds the parameters for the peer honesty handler
type PeerHonestyConfig struct {
	DecayCoefficient             float64
	DecayUpdateIntervalInSeconds uint32
	MaxScore                     float64
	MinScore                     float64
	BadPeerThreshold             float64
	UnitValue                    float64
}
