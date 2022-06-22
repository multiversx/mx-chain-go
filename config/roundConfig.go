package config

// RoundConfig contains round activation configurations
type RoundConfig struct {
	RoundActivations map[string]ActivationRoundByName
}

// ActivationRoundByName contains information related to a round activation event
type ActivationRoundByName struct {
	Round   string
	Options []string
}
