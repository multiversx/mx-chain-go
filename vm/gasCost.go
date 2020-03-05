package vm

// BaseOperationCost defines cost for base operation cost
type BaseOperationCost struct {
	StorePerByte    uint64
	ReleasePerByte  uint64
	DataCopyPerByte uint64
	PersistPerByte  uint64
	CompilePerByte  uint64
}

// MetaChainSystemSCsCost defines the cost of system staking SCs methods
type MetaChainSystemSCsCost struct {
	Stake               uint64
	UnStake             uint64
	UnBond              uint64
	Claim               uint64
	Get                 uint64
	ChangeRewardAddress uint64
	ChangeValidatorKeys uint64
	UnJail              uint64
}

// GasCost holds all the needed gas costs for system smart contracts
type GasCost struct {
	BaseOperationCost      BaseOperationCost
	MetaChainSystemSCsCost MetaChainSystemSCsCost
}
