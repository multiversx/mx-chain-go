package config

// RoundConfig contains round activation configurations
type RoundConfig struct {
	// Since we do not have yet any round based enable flag, this struct is empty for now.
	// Below, there is an example of how a new flag shall be added:
	// ActivationDummyFlag1 ActivationRoundByName
	// Where ActivationDummyFlag1 == table name from enableRounds.toml
}

// ActivationRoundByName contains information related to a round activation event
type ActivationRoundByName struct {
	Name  string
	Round uint64
}
