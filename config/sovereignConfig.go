package config

// SovereignConfig holds sovereign config
type SovereignConfig struct {
	ExtendedShardHdrNonceHashStorage StorageConfig
	ExtendedShardHeaderStorage       StorageConfig
	MainChainNotarization            MainChainNotarization `toml:"MainChainNotarization"`
}

type MainChainNotarization struct {
	StartRound uint64 `toml:"StartRound"`
}
