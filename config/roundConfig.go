package config

// RoundConfig contains round activation configurations
type RoundConfig struct {
	//TODO: Remove this once we have first activation flag
	ActivationDummy1 ActivationRoundByName
}

// ActivationRoundByName contains information related to a round activation event
type ActivationRoundByName struct {
	Name  string
	Round uint64
}
