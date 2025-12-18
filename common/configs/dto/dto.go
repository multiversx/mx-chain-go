package dto

type ConfigVariable string

const (
	NumFloodingRoundsFastReacting ConfigVariable = "NumFloodingRoundsFastReacting"
	NumFloodingRoundsSlowReacting ConfigVariable = "NumFloodingRoundsSlowReacting"
	NumFloodingRoundsOutOfSpecs   ConfigVariable = "NumFloodingRoundsOutOfSpecs"
)
