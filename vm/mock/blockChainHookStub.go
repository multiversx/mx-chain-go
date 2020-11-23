package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// BlockChainHookStub -
type BlockChainHookStub struct {
	AccountExtistsCalled          func(address []byte) (bool, error)
	NewAddressCalled              func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	GetStorageDataCalled          func(accountsAddress []byte, index []byte) ([]byte, error)
	GetUserAccountCalled          func(address []byte) (vmcommon.UserAccountHandler, error)
	GetShardOfAddressCalled       func(address []byte) uint32
	IsSmartContractCalled         func(address []byte) bool
	GetBlockHashCalled            func(nonce uint64) ([]byte, error)
	LastNonceCalled               func() uint64
	LastRoundCalled               func() uint64
	LastTimeStampCalled           func() uint64
	LastRandomSeedCalled          func() []byte
	LastEpochCalled               func() uint32
	GetStateRootHashCalled        func() []byte
	CurrentNonceCalled            func() uint64
	CurrentRoundCalled            func() uint64
	CurrentTimeStampCalled        func() uint64
	CurrentRandomSeedCalled       func() []byte
	CurrentEpochCalled            func() uint32
	ProcessBuiltInFunctionCalled  func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	GetBuiltinFunctionNamesCalled func() vmcommon.FunctionNames
	GetAllStateCalled             func(address []byte) (map[string][]byte, error)
	IsPayableCalled               func(address []byte) (bool, error)
}

// GetAccount -
func (b *BlockChainHookStub) AccountExists(address []byte) (bool, error) {
	if b.AccountExtistsCalled != nil {
		return b.AccountExtistsCalled(address)
	}
	return false, nil
}

// NewAddress -
func (b *BlockChainHookStub) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if b.NewAddressCalled != nil {
		return b.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}
	return []byte("newAddress"), nil
}

// GetStorageData -
func (b *BlockChainHookStub) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	if b.GetStorageDataCalled != nil {
		return b.GetStorageDataCalled(accountAddress, index)
	}
	return nil, nil
}

// GetUserAccount -
func (b *BlockChainHookStub) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	if b.GetUserAccountCalled != nil {
		return b.GetUserAccountCalled(address)
	}

	return state.NewUserAccount(address)
}

// GetShardOfAddress -
func (b *BlockChainHookStub) GetShardOfAddress(address []byte) uint32 {
	if b.GetShardOfAddressCalled != nil {
		return b.GetShardOfAddressCalled(address)
	}

	return 0
}

// IsSmartContract -
func (b *BlockChainHookStub) IsSmartContract(address []byte) bool {
	if b.IsSmartContractCalled != nil {
		return b.IsSmartContractCalled(address)
	}

	return false
}

// GetBlockhash -
func (b *BlockChainHookStub) GetBlockhash(nonce uint64) ([]byte, error) {
	if b.GetBlockHashCalled != nil {
		return b.GetBlockHashCalled(nonce)
	}
	return []byte("roothash"), nil
}

// LastNonce -
func (b *BlockChainHookStub) LastNonce() uint64 {
	if b.LastNonceCalled != nil {
		return b.LastNonceCalled()
	}
	return 0
}

// LastRound -
func (b *BlockChainHookStub) LastRound() uint64 {
	if b.LastRoundCalled != nil {
		return b.LastRoundCalled()
	}
	return 0
}

// LastTimeStamp -
func (b *BlockChainHookStub) LastTimeStamp() uint64 {
	if b.LastTimeStampCalled != nil {
		return b.LastTimeStampCalled()
	}
	return 0
}

// LastRandomSeed -
func (b *BlockChainHookStub) LastRandomSeed() []byte {
	if b.LastRandomSeedCalled != nil {
		return b.LastRandomSeedCalled()
	}
	return []byte("seed")
}

// LastEpoch -
func (b *BlockChainHookStub) LastEpoch() uint32 {
	if b.LastEpochCalled != nil {
		return b.LastEpochCalled()
	}
	return 0
}

// GetStateRootHash -
func (b *BlockChainHookStub) GetStateRootHash() []byte {
	if b.GetStateRootHashCalled != nil {
		return b.GetStateRootHashCalled()
	}
	return []byte("roothash")
}

// CurrentNonce -
func (b *BlockChainHookStub) CurrentNonce() uint64 {
	if b.CurrentNonceCalled != nil {
		return b.CurrentNonceCalled()
	}
	return 0
}

// CurrentRound -
func (b *BlockChainHookStub) CurrentRound() uint64 {
	if b.CurrentRoundCalled != nil {
		return b.CurrentRoundCalled()
	}
	return 0
}

// CurrentTimeStamp -
func (b *BlockChainHookStub) CurrentTimeStamp() uint64 {
	if b.CurrentTimeStampCalled != nil {
		return b.CurrentTimeStampCalled()
	}
	return 0
}

// CurrentRandomSeed -
func (b *BlockChainHookStub) CurrentRandomSeed() []byte {
	if b.CurrentRandomSeedCalled != nil {
		return b.CurrentRandomSeedCalled()
	}
	return []byte("seedseed")
}

// CurrentEpoch -
func (b *BlockChainHookStub) CurrentEpoch() uint32 {
	if b.CurrentEpochCalled != nil {
		return b.CurrentEpochCalled()
	}
	return 0
}

// ProcessBuiltInFunction -
func (b *BlockChainHookStub) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if b.ProcessBuiltInFunctionCalled != nil {
		return b.ProcessBuiltInFunctionCalled(input)
	}
	return &vmcommon.VMOutput{}, nil
}

// GetAllState -
func (b *BlockChainHookStub) GetAllState(address []byte) (map[string][]byte, error) {
	if b.GetAllStateCalled != nil {
		return b.GetAllStateCalled(address)
	}
	return nil, nil
}

// GetBuiltinFunctionNames -
func (b *BlockChainHookStub) GetBuiltinFunctionNames() vmcommon.FunctionNames {
	if b.GetBuiltinFunctionNamesCalled != nil {
		return b.GetBuiltinFunctionNamesCalled()
	}

	return make(vmcommon.FunctionNames)
}

// IsPayable -
func (b *BlockChainHookStub) IsPayable(address []byte) (bool, error) {
	if b.IsPayableCalled != nil {
		return b.IsPayableCalled(address)
	}

	return true, nil
}
