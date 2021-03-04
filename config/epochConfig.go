package config

// EpochConfig will hold epoch configurations
type EpochConfig struct {
	EnableEpochs EnableEpochs
}

// EnableEpochs will hold the configuration for activation epochs
type EnableEpochs struct {
	SCDeployEnableEpoch                    uint32
	BuiltInFunctionsEnableEpoch            uint32
	RelayedTransactionsEnableEpoch         uint32
	PenalizedTooMuchGasEnableEpoch         uint32
	SwitchJailWaitingEnableEpoch           uint32
	SwitchHysteresisForMinNodesEnableEpoch uint32
	BelowSignedThresholdEnableEpoch        uint32
	TransactionSignedWithTxHashEnableEpoch uint32
	MetaProtectionEnableEpoch              uint32
	AheadOfTimeGasUsageEnableEpoch         uint32
	GasPriceModifierEnableEpoch            uint32
	RepairCallbackEnableEpoch              uint32
	MaxNodesChangeEnableEpoch              []MaxNodesChangeConfig
	BlockGasAndFeesReCheckEnableEpoch      uint32
	StakingV2Epoch                         uint32
	StakeEnableEpoch                       uint32
	DoubleKeyProtectionEnableEpoch         uint32
	ESDTEnableEpoch                        uint32
	GovernanceEnableEpoch                  uint32
	DelegationManagerEnableEpoch           uint32
	DelegationSmartContractEnableEpoch     uint32
}
