package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// BlockChainHookStub -
type BlockChainHookStub struct {
	NewAddressCalled                     func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	GetStorageDataCalled                 func(accountsAddress []byte, index []byte) ([]byte, uint32, error)
	GetBlockHashCalled                   func(nonce uint64) ([]byte, error)
	LastNonceCalled                      func() uint64
	LastRoundCalled                      func() uint64
	LastTimeStampCalled                  func() uint64
	LastRandomSeedCalled                 func() []byte
	LastEpochCalled                      func() uint32
	GetStateRootHashCalled               func() []byte
	CurrentNonceCalled                   func() uint64
	CurrentRoundCalled                   func() uint64
	CurrentTimeStampCalled               func() uint64
	CurrentRandomSeedCalled              func() []byte
	CurrentEpochCalled                   func() uint32
	ProcessBuiltInFunctionCalled         func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	GetBuiltinFunctionNamesCalled        func() vmcommon.FunctionNames
	GetBuiltinFunctionsContainerCalled   func() vmcommon.BuiltInFunctionContainer
	GetAllStateCalled                    func(address []byte) (map[string][]byte, error)
	GetUserAccountCalled                 func(address []byte) (vmcommon.UserAccountHandler, error)
	GetShardOfAddressCalled              func(address []byte) uint32
	IsSmartContractCalled                func(address []byte) bool
	IsPayableCalled                      func(sndAddress []byte, recvAddress []byte) (bool, error)
	GetCompiledCodeCalled                func(codeHash []byte) (bool, []byte)
	SaveCompiledCodeCalled               func(codeHash []byte, code []byte)
	ClearCompiledCodesCalled             func()
	GetCodeCalled                        func(account vmcommon.UserAccountHandler) []byte
	GetESDTTokenCalled                   func(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	NumberOfShardsCalled                 func() uint32
	GetSnapshotCalled                    func() int
	RevertToSnapshotCalled               func(snapshot int) error
	SetCurrentHeaderCalled               func(hdr data.HeaderHandler)
	DeleteCompiledCodeCalled             func(codeHash []byte)
	SaveNFTMetaDataToSystemAccountCalled func(tx data.TransactionHandler) error
	CloseCalled                          func() error
	FilterCodeMetadataForUpgradeCalled   func(input []byte) ([]byte, error)
	ApplyFiltersOnCodeMetadataCalled     func(codeMetadata vmcommon.CodeMetadata) vmcommon.CodeMetadata
	ResetCountersCalled                  func()
	GetCounterValuesCalled               func() map[string]uint64
}

// GetCode -
func (stub *BlockChainHookStub) GetCode(account vmcommon.UserAccountHandler) []byte {
	if stub.GetCodeCalled != nil {
		return stub.GetCodeCalled(account)
	}

	return make([]byte, 0)
}

// GetUserAccount -
func (stub *BlockChainHookStub) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	if stub.GetUserAccountCalled != nil {
		return stub.GetUserAccountCalled(address)
	}

	return nil, nil
}

// GetStorageData -
func (stub *BlockChainHookStub) GetStorageData(accountAddress []byte, index []byte) ([]byte, uint32, error) {
	if stub.GetStorageDataCalled != nil {
		return stub.GetStorageDataCalled(accountAddress, index)
	}

	return make([]byte, 0), 0, nil
}

// GetBlockhash -
func (stub *BlockChainHookStub) GetBlockhash(nonce uint64) ([]byte, error) {
	if stub.GetBlockHashCalled != nil {
		return stub.GetBlockHashCalled(nonce)
	}

	return make([]byte, 0), nil
}

// LastNonce -
func (stub *BlockChainHookStub) LastNonce() uint64 {
	if stub.LastNonceCalled != nil {
		return stub.LastNonceCalled()
	}

	return 0
}

// LastRound -
func (stub *BlockChainHookStub) LastRound() uint64 {
	if stub.LastRoundCalled != nil {
		return stub.LastRoundCalled()
	}

	return 0
}

// LastTimeStamp -
func (stub *BlockChainHookStub) LastTimeStamp() uint64 {
	if stub.LastTimeStampCalled != nil {
		return stub.LastTimeStampCalled()
	}

	return 0
}

// LastRandomSeed -
func (stub *BlockChainHookStub) LastRandomSeed() []byte {
	if stub.LastRandomSeedCalled != nil {
		return stub.LastRandomSeedCalled()
	}

	return make([]byte, 0)
}

// LastEpoch -
func (stub *BlockChainHookStub) LastEpoch() uint32 {
	if stub.LastEpochCalled != nil {
		return stub.LastEpochCalled()
	}

	return 0
}

// GetStateRootHash -
func (stub *BlockChainHookStub) GetStateRootHash() []byte {
	if stub.GetStateRootHashCalled != nil {
		return stub.GetStateRootHashCalled()
	}

	return make([]byte, 0)
}

// CurrentNonce -
func (stub *BlockChainHookStub) CurrentNonce() uint64 {
	if stub.CurrentNonceCalled != nil {
		return stub.CurrentNonceCalled()
	}

	return 0
}

// CurrentRound -
func (stub *BlockChainHookStub) CurrentRound() uint64 {
	if stub.CurrentRoundCalled != nil {
		return stub.CurrentRoundCalled()
	}

	return 0
}

// CurrentTimeStamp -
func (stub *BlockChainHookStub) CurrentTimeStamp() uint64 {
	if stub.CurrentTimeStampCalled != nil {
		return stub.CurrentTimeStampCalled()
	}

	return 0
}

// CurrentRandomSeed -
func (stub *BlockChainHookStub) CurrentRandomSeed() []byte {
	if stub.CurrentRandomSeedCalled != nil {
		return stub.CurrentRandomSeedCalled()
	}

	return make([]byte, 0)
}

// CurrentEpoch -
func (stub *BlockChainHookStub) CurrentEpoch() uint32 {
	if stub.CurrentEpochCalled != nil {
		return stub.CurrentEpochCalled()
	}

	return 0
}

// NewAddress -
func (stub *BlockChainHookStub) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if stub.NewAddressCalled != nil {
		return stub.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}

	return make([]byte, 0), nil
}

// ProcessBuiltInFunction -
func (stub *BlockChainHookStub) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if stub.ProcessBuiltInFunctionCalled != nil {
		return stub.ProcessBuiltInFunctionCalled(input)
	}

	return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
}

// SaveNFTMetaDataToSystemAccount -
func (stub *BlockChainHookStub) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	if stub.SaveNFTMetaDataToSystemAccountCalled != nil {
		return stub.SaveNFTMetaDataToSystemAccountCalled(tx)
	}

	return nil
}

// GetShardOfAddress -
func (stub *BlockChainHookStub) GetShardOfAddress(address []byte) uint32 {
	if stub.GetShardOfAddressCalled != nil {
		return stub.GetShardOfAddressCalled(address)
	}

	return 0
}

// IsSmartContract -
func (stub *BlockChainHookStub) IsSmartContract(address []byte) bool {
	if stub.IsSmartContractCalled(address) {
		return stub.IsSmartContractCalled(address)
	}

	return false
}

// GetBuiltinFunctionNames -
func (stub *BlockChainHookStub) GetBuiltinFunctionNames() vmcommon.FunctionNames {
	if stub.GetBuiltinFunctionNamesCalled != nil {
		return stub.GetBuiltinFunctionNamesCalled()
	}

	return make(map[string]struct{})
}

// GetBuiltinFunctionsContainer -
func (stub *BlockChainHookStub) GetBuiltinFunctionsContainer() vmcommon.BuiltInFunctionContainer {
	if stub.GetBuiltinFunctionsContainerCalled != nil {
		return stub.GetBuiltinFunctionsContainerCalled()
	}

	return nil
}

// GetAllState -
func (stub *BlockChainHookStub) GetAllState(address []byte) (map[string][]byte, error) {
	if stub.GetAllStateCalled != nil {
		return stub.GetAllStateCalled(address)
	}

	return nil, nil
}

// GetESDTToken -
func (stub *BlockChainHookStub) GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error) {
	if stub.GetESDTTokenCalled != nil {
		return stub.GetESDTTokenCalled(address, tokenID, nonce)
	}

	return &esdt.ESDigitalToken{}, nil
}

// NumberOfShards -
func (stub *BlockChainHookStub) NumberOfShards() uint32 {
	if stub.NumberOfShardsCalled != nil {
		return stub.NumberOfShardsCalled()
	}

	return 0
}

// SetCurrentHeader -
func (stub *BlockChainHookStub) SetCurrentHeader(hdr data.HeaderHandler) {
	if stub.SetCurrentHeaderCalled != nil {
		stub.SetCurrentHeaderCalled(hdr)
	}
}

// SaveCompiledCode -
func (stub *BlockChainHookStub) SaveCompiledCode(codeHash []byte, code []byte) {
	if stub.SaveCompiledCodeCalled != nil {
		stub.SaveCompiledCodeCalled(codeHash, code)
	}
}

// GetCompiledCode -
func (stub *BlockChainHookStub) GetCompiledCode(codeHash []byte) (bool, []byte) {
	if stub.GetCompiledCodeCalled != nil {
		return stub.GetCompiledCodeCalled(codeHash)
	}

	return false, nil
}

// IsPayable -
func (stub *BlockChainHookStub) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	if stub.IsPayableCalled != nil {
		return stub.IsPayableCalled(sndAddress, recvAddress)
	}

	return true, nil
}

// DeleteCompiledCode -
func (stub *BlockChainHookStub) DeleteCompiledCode(codeHash []byte) {
	if stub.DeleteCompiledCodeCalled != nil {
		stub.DeleteCompiledCodeCalled(codeHash)
	}
}

// ClearCompiledCodes -
func (stub *BlockChainHookStub) ClearCompiledCodes() {
	if stub.ClearCompiledCodesCalled != nil {
		stub.ClearCompiledCodesCalled()
	}
}

// GetSnapshot -
func (stub *BlockChainHookStub) GetSnapshot() int {
	if stub.GetSnapshotCalled != nil {
		return stub.GetSnapshotCalled()
	}

	return 0
}

// RevertToSnapshot -
func (stub *BlockChainHookStub) RevertToSnapshot(snapshot int) error {
	if stub.RevertToSnapshotCalled != nil {
		return stub.RevertToSnapshotCalled(snapshot)
	}

	return nil
}

// FilterCodeMetadataForUpgrade -
func (stub *BlockChainHookStub) FilterCodeMetadataForUpgrade(input []byte) ([]byte, error) {
	if stub.FilterCodeMetadataForUpgradeCalled != nil {
		return stub.FilterCodeMetadataForUpgradeCalled(input)
	}

	return input, nil
}

// ApplyFiltersOnSCCodeMetadata -
func (stub *BlockChainHookStub) ApplyFiltersOnSCCodeMetadata(codeMetadata vmcommon.CodeMetadata) vmcommon.CodeMetadata {
	if stub.ApplyFiltersOnCodeMetadataCalled != nil {
		stub.ApplyFiltersOnCodeMetadataCalled(codeMetadata)
	}

	return codeMetadata
}

// IsPaused -
func (stub *BlockChainHookStub) IsPaused(_ []byte) bool {
	return false
}

// IsLimitedTransfer -
func (stub *BlockChainHookStub) IsLimitedTransfer(_ []byte) bool {
	return false
}

// Close -
func (stub *BlockChainHookStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// ResetCounters -
func (stub *BlockChainHookStub) ResetCounters() {
	if stub.ResetCountersCalled != nil {
		stub.ResetCountersCalled()
	}
}

// GetCounterValues -
func (stub *BlockChainHookStub) GetCounterValues() map[string]uint64 {
	if stub.GetCounterValuesCalled != nil {
		return stub.GetCounterValuesCalled()
	}

	return make(map[string]uint64)
}

// IsInterfaceNil -
func (stub *BlockChainHookStub) IsInterfaceNil() bool {
	return stub == nil
}
