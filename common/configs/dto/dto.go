package dto

// ConfigVariable defines a config variable name
type ConfigVariable string

const (
	// NumFloodingRoundsFastReacting defines variable name for NumFloodingRoundsFastReacting
	NumFloodingRoundsFastReacting ConfigVariable = "NumFloodingRoundsFastReacting"
	// NumFloodingRoundsSlowReacting defines variable name for NumFloodingRoundsSlowReacting
	NumFloodingRoundsSlowReacting ConfigVariable = "NumFloodingRoundsSlowReacting"
	// NumFloodingRoundsOutOfSpecs defines variable name for NumFloodingRoundsOutOfSpecs
	NumFloodingRoundsOutOfSpecs ConfigVariable = "NumFloodingRoundsOutOfSpecs"
	// MaxConsecutiveRoundsOfRatingDecrease defines variable name for MaxConsecutiveRoundsOfRatingDecrease
	MaxConsecutiveRoundsOfRatingDecrease ConfigVariable = "MaxConsecutiveRoundsOfRatingDecrease"
)
