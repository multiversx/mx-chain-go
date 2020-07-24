package config

// SystemSmartContractsConfig defines the system smart contract configs
type SystemSmartContractsConfig struct {
	ESDTSystemSCConfig       ESDTSystemSCConfig
	GovernanceSystemSCConfig GovernanceSystemSCConfig
}

// ESDTSystemSCConfig defines a set of constant to initialize the esdt system smart contract
type ESDTSystemSCConfig struct {
	BaseIssuingCost string
	OwnerAddress    string
	Disabled        bool
}

// GovernanceSystemSCConfig defines the set of constants to initialize the governance system smart contract
type GovernanceSystemSCConfig struct {
	ProposalCost     string
	NumNodes         int64
	MinQuorum        int32
	MinPassThreshold int32
	MinVetoThreshold int32
	Disabled         bool
}
