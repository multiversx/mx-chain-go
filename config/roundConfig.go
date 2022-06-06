package config

// RoundConfig contains round activation configurations
type RoundConfig struct {
	RoundActivations []ActivationRoundByName
}

// ActivationRoundByName contains information related to a round activation event
type ActivationRoundByName struct {
	Name    string
	Round   uint64
	Options []string
}
