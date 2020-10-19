package config

// SystemSmartContractsConfig defines the system smart contract configs
type SystemSmartContractsConfig struct {
	ESDTSystemSCConfig              ESDTSystemSCConfig
	GovernanceSystemSCConfig        GovernanceSystemSCConfig
	StakingSystemSCConfig           StakingSystemSCConfig
	DelegationManagerSystemSCConfig DelegationManagerSystemSCConfig
	DelegationSystemSCConfig        DelegationSystemSCConfig
}

// StakingSystemSCConfig will hold the staking system smart contract settings
type StakingSystemSCConfig struct {
	GenesisNodePrice                     string
	MinStakeValue                        string
	MinUnstakeTokensValue                string
	UnJailValue                          string
	MinStepValue                         string
	UnBondPeriod                         uint64
	StakingV2Epoch                       uint32
	StakeEnableEpoch                     uint32
	NumRoundsWithoutBleed                uint64
	MaximumPercentageToBleed             float64
	BleedPercentagePerRound              float64
	MaxNumberOfNodesForStake             uint64
	NodesToSelectInAuction               uint64
	ActivateBLSPubKeyMessageVerification bool
}

// ESDTSystemSCConfig defines a set of constant to initialize the esdt system smart contract
type ESDTSystemSCConfig struct {
	BaseIssuingCost string
	OwnerAddress    string
	EnabledEpoch    uint32
}

// GovernanceSystemSCConfig defines the set of constants to initialize the governance system smart contract
type GovernanceSystemSCConfig struct {
	ProposalCost     string
	NumNodes         int64
	MinQuorum        int32
	MinPassThreshold int32
	MinVetoThreshold int32
	EnabledEpoch     uint32
}

// DelegationManagerSystemSCConfig defines a set of constants to initialize the delegation manager system smart contract
type DelegationManagerSystemSCConfig struct {
	BaseIssuingCost    string
	MinCreationDeposit string
	EnabledEpoch       uint32
}

// DelegationSystemSCConfig defines a set of constants to initialize the delegation system smart contract
type DelegationSystemSCConfig struct {
	MinStakeAmount string
	EnabledEpoch   uint32
	MinServiceFee  uint64
	MaxServiceFee  uint64
}
