package config

// TrieOperationsConfig bounds non-paginated API trie collection. Zero fields use defaults.
type TrieOperationsConfig struct {
	MaxBytes                uint64
	MaxLeaves               uint64
	MaxConcurrentRequests   uint32
	HTTPTimeoutMilliseconds uint32
	MaxRequestBodyBytes     uint32
}

// WithDefaults supplies limits even when web antiflood is disabled.
func (cfg TrieOperationsConfig) WithDefaults() TrieOperationsConfig {
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 10 * 1024 * 1024
	}
	if cfg.MaxLeaves == 0 {
		cfg.MaxLeaves = 10000
	}
	if cfg.MaxConcurrentRequests == 0 {
		cfg.MaxConcurrentRequests = 2
	}
	if cfg.HTTPTimeoutMilliseconds == 0 {
		cfg.HTTPTimeoutMilliseconds = 5000
	}
	if cfg.MaxRequestBodyBytes == 0 {
		cfg.MaxRequestBodyBytes = 2 * 1024 * 1024
	}
	return cfg
}
