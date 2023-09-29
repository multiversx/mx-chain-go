package config

// SovereignConfig holds sovereign config
type SovereignConfig struct {
	ExtendedShardHdrNonceHashStorage StorageConfig
	ExtendedShardHeaderStorage       StorageConfig
	MainChainNotarization            MainChainNotarization `toml:"MainChainNotarization"`
}

// MainChainNotarization defines necessary data to start main chain notarization on a sovereign shard
type MainChainNotarization struct {
	MainChainNotarizationStartRound uint64 `toml:"MainChainNotarizationStartRound"`
}
