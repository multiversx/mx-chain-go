package vm

// BaseOperationCost defines cost for base operation cost
type BaseOperationCost struct {
	StorePerByte      uint64
	ReleasePerByte    uint64
	DataCopyPerByte   uint64
	PersistPerByte    uint64
	CompilePerByte    uint64
	AoTPreparePerByte uint64
	GetCode           uint64
}

// MetaChainSystemSCsCost defines the cost of system staking SCs methods
type MetaChainSystemSCsCost struct {
	Stake                 uint64
	UnStake               uint64
	UnBond                uint64
	Claim                 uint64
	Get                   uint64
	ChangeRewardAddress   uint64
	ChangeValidatorKeys   uint64
	UnJail                uint64
	ESDTIssue             uint64
	ESDTOperations        uint64
	Proposal              uint64
	Vote                  uint64
	DelegateVote          uint64
	RevokeVote            uint64
	CloseProposal         uint64
	ClearProposal         uint64
	ClaimAccumulatedFees  uint64
	ChangeConfig          uint64
	DelegationOps         uint64
	UnStakeTokens         uint64
	UnBondTokens          uint64
	DelegationMgrOps      uint64
	ValidatorToDelegation uint64
	GetAllNodeStates      uint64
	GetActiveFund         uint64
	FixWaitingListSize    uint64
}

// BuiltInCost defines cost for built-in methods
type BuiltInCost struct {
	ChangeOwnerAddress       uint64
	ClaimDeveloperRewards    uint64
	SaveUserName             uint64
	SaveKeyValue             uint64
	ESDTTransfer             uint64
	ESDTBurn                 uint64
	ESDTLocalMint            uint64
	ESDTLocalBurn            uint64
	ESDTNFTCreate            uint64
	ESDTNFTAddQuantity       uint64
	ESDTNFTBurn              uint64
	ESDTNFTTransfer          uint64
	ESDTNFTChangeCreateOwner uint64
	ESDTNFTAddUri            uint64
	ESDTNFTUpdateAttributes  uint64
	ESDTNFTMultiTransfer     uint64
	TrieLoadPerNode          uint64
	TrieStorePerNode         uint64
}

// GasCost holds all the needed gas costs for system smart contracts
type GasCost struct {
	BaseOperationCost      BaseOperationCost
	MetaChainSystemSCsCost MetaChainSystemSCsCost
	BuiltInCost            BuiltInCost
}
