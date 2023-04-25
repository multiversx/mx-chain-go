package process

// BaseOperationCost defines cost for base operation cost
type BaseOperationCost struct {
	StorePerByte      uint64
	ReleasePerByte    uint64
	DataCopyPerByte   uint64
	PersistPerByte    uint64
	CompilePerByte    uint64
	AoTPreparePerByte uint64
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
	SetGuardian              uint64
	GuardAccount             uint64
	UnGuardAccount           uint64
	TrieLoadPerNode          uint64
	TrieStorePerNode         uint64
}

// GasCost holds all the needed gas costs for system smart contracts
type GasCost struct {
	BaseOperationCost BaseOperationCost
	BuiltInCost       BuiltInCost
}
