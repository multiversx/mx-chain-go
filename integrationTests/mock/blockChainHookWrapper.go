package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// BlockChainHookWrapper -
type BlockChainHookWrapper struct {
	BlockchainHook vmcommon.BlockchainHook

	NewAddressCalled                     func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	GetStorageDataCalled                 func(accountsAddress []byte, index []byte) ([]byte, error)
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
	GetAllStateCalled                    func(address []byte) (map[string][]byte, error)
	GetUserAccountCalled                 func(address []byte) (vmcommon.UserAccountHandler, error)
	GetShardOfAddressCalled              func(address []byte) uint32
	IsSmartContractCalled                func(address []byte) bool
	IsPayableCalled                      func(address []byte) (bool, error)
	GetCompiledCodeCalled                func(codeHash []byte) (bool, []byte)
	SaveCompiledCodeCalled               func(codeHash []byte, code []byte)
	ClearCompiledCodesCalled             func()
	GetCodeCalled                        func(account vmcommon.UserAccountHandler) []byte
	GetESDTTokenCalled                   func(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error)
	GetSnapshotCalled                    func() int
	RevertToSnapshotCalled               func(snapshot int) error
	SetCurrentHeaderCalled               func(hdr data.HeaderHandler)
	DeleteCompiledCodeCalled             func(codeHash []byte)
	SaveNFTMetaDataToSystemAccountCalled func(tx data.TransactionHandler) error
}

// NewBlockChainHookWrapper -
func NewBlockChainHookWrapper(blockchainHook vmcommon.BlockchainHook) *BlockChainHookWrapper {
	return &BlockChainHookWrapper{
		BlockchainHook: blockchainHook,
	}
}

// NewAddress -
func (wrapper *BlockChainHookWrapper) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if wrapper.NewAddressCalled != nil {
		return wrapper.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}

	return wrapper.BlockchainHook.NewAddress(creatorAddress, creatorNonce, vmType)
}

// GetStorageData -
func (wrapper *BlockChainHookWrapper) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	if wrapper.GetStorageDataCalled != nil {
		return wrapper.GetStorageDataCalled(accountAddress, index)
	}

	return wrapper.BlockchainHook.GetStorageData(accountAddress, index)
}

// GetBlockhash -
func (wrapper *BlockChainHookWrapper) GetBlockhash(nonce uint64) ([]byte, error) {
	if wrapper.GetBlockHashCalled != nil {
		return wrapper.GetBlockHashCalled(nonce)
	}

	return wrapper.BlockchainHook.GetBlockhash(nonce)
}

// LastNonce -
func (wrapper *BlockChainHookWrapper) LastNonce() uint64 {
	if wrapper.LastNonceCalled != nil {
		return wrapper.LastNonceCalled()
	}

	return wrapper.BlockchainHook.LastNonce()
}

// LastRound -
func (wrapper *BlockChainHookWrapper) LastRound() uint64 {
	if wrapper.LastRoundCalled != nil {
		return wrapper.LastRoundCalled()
	}

	return wrapper.BlockchainHook.LastRound()
}

// LastTimeStamp -
func (wrapper *BlockChainHookWrapper) LastTimeStamp() uint64 {
	if wrapper.LastTimeStampCalled != nil {
		return wrapper.LastTimeStampCalled()
	}

	return wrapper.BlockchainHook.LastTimeStamp()
}

// LastRandomSeed -
func (wrapper *BlockChainHookWrapper) LastRandomSeed() []byte {
	if wrapper.LastRandomSeedCalled != nil {
		return wrapper.LastRandomSeedCalled()
	}

	return wrapper.BlockchainHook.LastRandomSeed()
}

// LastEpoch -
func (wrapper *BlockChainHookWrapper) LastEpoch() uint32 {
	if wrapper.LastEpochCalled != nil {
		return wrapper.LastEpochCalled()
	}

	return wrapper.BlockchainHook.LastEpoch()
}

// GetStateRootHash -
func (wrapper *BlockChainHookWrapper) GetStateRootHash() []byte {
	if wrapper.GetStateRootHashCalled != nil {
		return wrapper.GetStateRootHashCalled()
	}

	return wrapper.BlockchainHook.GetStateRootHash()
}

// CurrentNonce -
func (wrapper *BlockChainHookWrapper) CurrentNonce() uint64 {
	if wrapper.CurrentNonceCalled != nil {
		return wrapper.CurrentNonceCalled()
	}

	return wrapper.BlockchainHook.CurrentNonce()
}

// CurrentRound -
func (wrapper *BlockChainHookWrapper) CurrentRound() uint64 {
	if wrapper.CurrentRoundCalled != nil {
		return wrapper.CurrentRoundCalled()
	}

	return wrapper.BlockchainHook.CurrentRound()
}

// CurrentTimeStamp -
func (wrapper *BlockChainHookWrapper) CurrentTimeStamp() uint64 {
	if wrapper.CurrentTimeStampCalled != nil {
		return wrapper.CurrentTimeStampCalled()
	}

	return wrapper.BlockchainHook.CurrentTimeStamp()
}

// CurrentRandomSeed -
func (wrapper *BlockChainHookWrapper) CurrentRandomSeed() []byte {
	if wrapper.CurrentRandomSeedCalled != nil {
		return wrapper.CurrentRandomSeedCalled()
	}

	return wrapper.BlockchainHook.CurrentRandomSeed()
}

// CurrentEpoch -
func (wrapper *BlockChainHookWrapper) CurrentEpoch() uint32 {
	if wrapper.CurrentEpochCalled != nil {
		return wrapper.CurrentEpochCalled()
	}

	return wrapper.BlockchainHook.CurrentEpoch()
}

// ProcessBuiltInFunction -
func (wrapper *BlockChainHookWrapper) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if wrapper.ProcessBuiltInFunctionCalled != nil {
		return wrapper.ProcessBuiltInFunctionCalled(input)
	}

	return wrapper.BlockchainHook.ProcessBuiltInFunction(input)
}

// GetBuiltinFunctionNames -
func (wrapper *BlockChainHookWrapper) GetBuiltinFunctionNames() vmcommon.FunctionNames {
	if wrapper.GetBuiltinFunctionNamesCalled != nil {
		return wrapper.GetBuiltinFunctionNamesCalled()
	}

	return wrapper.BlockchainHook.GetBuiltinFunctionNames()
}

// GetAllState -
func (wrapper *BlockChainHookWrapper) GetAllState(address []byte) (map[string][]byte, error) {
	if wrapper.GetAllStateCalled != nil {
		return wrapper.GetAllStateCalled(address)
	}

	return wrapper.BlockchainHook.GetAllState(address)
}

// GetUserAccount -
func (wrapper *BlockChainHookWrapper) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	if wrapper.GetUserAccountCalled != nil {
		return wrapper.GetUserAccountCalled(address)
	}

	return wrapper.BlockchainHook.GetUserAccount(address)
}

// GetESDTToken -
func (wrapper *BlockChainHookWrapper) GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error) {
	if wrapper.GetESDTTokenCalled != nil {
		return wrapper.GetESDTTokenCalled(address, tokenID, nonce)
	}

	return wrapper.BlockchainHook.GetESDTToken(address, tokenID, nonce)
}

// GetCode -
func (wrapper *BlockChainHookWrapper) GetCode(account vmcommon.UserAccountHandler) []byte {
	if wrapper.GetCodeCalled != nil {
		return wrapper.GetCodeCalled(account)
	}

	return wrapper.BlockchainHook.GetCode(account)
}

// GetShardOfAddress -
func (wrapper *BlockChainHookWrapper) GetShardOfAddress(address []byte) uint32 {
	if wrapper.GetShardOfAddressCalled != nil {
		return wrapper.GetShardOfAddressCalled(address)
	}

	return wrapper.BlockchainHook.GetShardOfAddress(address)
}

// IsSmartContract -
func (wrapper *BlockChainHookWrapper) IsSmartContract(address []byte) bool {
	if wrapper.IsSmartContractCalled != nil {
		return wrapper.IsSmartContractCalled(address)
	}

	return wrapper.BlockchainHook.IsSmartContract(address)
}

// IsPayable -
func (wrapper *BlockChainHookWrapper) IsPayable(senderAddress, receiverAddress []byte) (bool, error) {
	if wrapper.IsPayableCalled != nil {
		return wrapper.IsPayableCalled(receiverAddress)
	}

	return wrapper.BlockchainHook.IsPayable(senderAddress, receiverAddress)
}

// SaveCompiledCode -
func (wrapper *BlockChainHookWrapper) SaveCompiledCode(codeHash []byte, code []byte) {
	if wrapper.SaveCompiledCodeCalled != nil {
		wrapper.SaveCompiledCodeCalled(codeHash, code)
		return
	}

	wrapper.BlockchainHook.SaveCompiledCode(codeHash, code)
}

// GetCompiledCode -
func (wrapper *BlockChainHookWrapper) GetCompiledCode(codeHash []byte) (bool, []byte) {
	if wrapper.GetCompiledCodeCalled != nil {
		return wrapper.GetCompiledCodeCalled(codeHash)
	}

	return wrapper.BlockchainHook.GetCompiledCode(codeHash)
}

// ClearCompiledCodes -
func (wrapper *BlockChainHookWrapper) ClearCompiledCodes() {
	if wrapper.ClearCompiledCodesCalled != nil {
		wrapper.ClearCompiledCodesCalled()
		return
	}

	wrapper.BlockchainHook.ClearCompiledCodes()
}

// GetSnapshot -
func (wrapper *BlockChainHookWrapper) GetSnapshot() int {
	if wrapper.GetSnapshotCalled != nil {
		return wrapper.GetSnapshotCalled()
	}

	return wrapper.BlockchainHook.GetSnapshot()
}

// RevertToSnapshot -
func (wrapper *BlockChainHookWrapper) RevertToSnapshot(snapshot int) error {
	if wrapper.RevertToSnapshotCalled != nil {
		return wrapper.RevertToSnapshotCalled(snapshot)
	}

	return wrapper.BlockchainHook.RevertToSnapshot(snapshot)
}

// SetCurrentHeader -
func (wrapper *BlockChainHookWrapper) SetCurrentHeader(hdr data.HeaderHandler) {
	if wrapper.SetCurrentHeaderCalled != nil {
		wrapper.SetCurrentHeaderCalled(hdr)
		return
	}

	wrapper.BlockchainHook.(process.BlockChainHookHandler).SetCurrentHeader(hdr)
}

// DeleteCompiledCode -
func (wrapper *BlockChainHookWrapper) DeleteCompiledCode(codeHash []byte) {
	if wrapper.DeleteCompiledCodeCalled != nil {
		wrapper.DeleteCompiledCodeCalled(codeHash)
		return
	}

	wrapper.BlockchainHook.(process.BlockChainHookHandler).DeleteCompiledCode(codeHash)
}

// SaveNFTMetaDataToSystemAccount -
func (wrapper *BlockChainHookWrapper) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	if wrapper.SaveNFTMetaDataToSystemAccountCalled != nil {
		return wrapper.SaveNFTMetaDataToSystemAccountCalled(tx)
	}

	return wrapper.BlockchainHook.(process.BlockChainHookHandler).SaveNFTMetaDataToSystemAccount(tx)
}

// IsInterfaceNil -
func (wrapper *BlockChainHookWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
