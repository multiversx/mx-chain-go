package config

// SystemSmartContractsConfig defines the system smart contract configs
type SystemSmartContractsConfig struct {
	ESDTSystemSCConfig              ESDTSystemSCConfig
	GovernanceSystemSCConfig        GovernanceSystemSCConfig
	StakingSystemSCConfig           StakingSystemSCConfig
	DelegationManagerSystemSCConfig DelegationManagerSystemSCConfig
	DelegationSystemSCConfig        DelegationSystemSCConfig
	SoftAuctionConfig               SoftAuctionConfig
}

// StakingSystemSCConfig will hold the staking system smart contract settings
type StakingSystemSCConfig struct {
	GenesisNodePrice                     string
	MinStakeValue                        string
	MinUnstakeTokensValue                string
	UnJailValue                          string
	MinStepValue                         string
	UnBondPeriod                         uint64
	UnBondPeriodInEpochs                 uint32
	NumRoundsWithoutBleed                uint64
	MaximumPercentageToBleed             float64
	BleedPercentagePerRound              float64
	MaxNumberOfNodesForStake             uint64
	ActivateBLSPubKeyMessageVerification bool
	StakeLimitPercentage                 float64
	NodeLimitPercentage                  float64
}

// ESDTSystemSCConfig defines a set of constant to initialize the esdt system smart contract
type ESDTSystemSCConfig struct {
	BaseIssuingCost string
	OwnerAddress    string
}

// GovernanceSystemSCConfigV1 holds the initial set of values that were used to initialise the
// governance system smart contract at genesis time
type GovernanceSystemSCConfigV1 struct {
	NumNodes         int64
	ProposalCost     string
	MinQuorum        int32
	MinPassThreshold int32
	MinVetoThreshold int32
}

// GovernanceSystemSCConfigActive defines the set of configuration values used by the governance
// system smart contract once it activates
type GovernanceSystemSCConfigActive struct {
	ProposalCost     string
	LostProposalFee  string
	MinQuorum        float64
	MinPassThreshold float64
	MinVetoThreshold float64
}

// GovernanceSystemSCConfig defines the set of constants to initialize the governance system smart contract
type GovernanceSystemSCConfig struct {
	V1           GovernanceSystemSCConfigV1
	Active       GovernanceSystemSCConfigActive
	OwnerAddress string
}

// DelegationManagerSystemSCConfig defines a set of constants to initialize the delegation manager system smart contract
type DelegationManagerSystemSCConfig struct {
	MinCreationDeposit  string
	MinStakeAmount      string
	ConfigChangeAddress string
}

// DelegationSystemSCConfig defines a set of constants to initialize the delegation system smart contract
type DelegationSystemSCConfig struct {
	MinServiceFee               uint64
	MaxServiceFee               uint64
	AddTokensWhitelistedAddress string
}

// SoftAuctionConfig represents the config options for soft auction selecting used in staking v4
type SoftAuctionConfig struct {
	TopUpStep             string
	MinTopUp              string
	MaxTopUp              string
	MaxNumberOfIterations uint64
}
