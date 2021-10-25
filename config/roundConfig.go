package config

// RoundConfig contains round activation configurations
type RoundConfig struct {
	EnableRoundsByName []EnableRoundsByName
}

// EnableRoundsByName contains information related to a round activation event
type EnableRoundsByName struct {
	Name  string
	Round uint64
	Shard uint32
}
