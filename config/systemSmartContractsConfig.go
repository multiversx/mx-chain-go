package config

// SystemSmartContractsConfig defines the system smart contract configs
type SystemSmartContractsConfig struct {
	ESDTSystemSCConfig ESDTSystemSCConfig
}

// ESDTSystemSCConfig defines a set of constant to initialize the esdt system smart contract
type ESDTSystemSCConfig struct {
	BaseIssuingCost string
	OwnerAddress    string
}
