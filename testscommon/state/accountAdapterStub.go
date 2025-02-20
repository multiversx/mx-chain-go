package state

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
)

// StateUserAccountHandlerStub -
type StateUserAccountHandlerStub struct {
	AddressBytesCalled          func() []byte
	IncreaseNonceCalled         func(nonce uint64)
	GetNonceCalled              func() uint64
	SetCodeCalled               func(code []byte)
	SetCodeMetadataCalled       func(codeMetadata []byte)
	GetCodeMetadataCalled       func() []byte
	SetCodeHashCalled           func([]byte)
	GetCodeHashCalled           func() []byte
	SetRootHashCalled           func([]byte)
	GetRootHashCalled           func() []byte
	SetDataTrieCalled           func(trie common.Trie)
	DataTrieCalled              func() common.DataTrieHandler
	RetrieveValueCalled         func(key []byte) ([]byte, uint32, error)
	SaveKeyValueCalled          func(key []byte, value []byte) error
	AddToBalanceCalled          func(value *big.Int) error
	SubFromBalanceCalled        func(value *big.Int) error
	GetBalanceCalled            func() *big.Int
	ClaimDeveloperRewardsCalled func([]byte) (*big.Int, error)
	AddToDeveloperRewardCalled  func(*big.Int)
	GetDeveloperRewardCalled    func() *big.Int
	ChangeOwnerAddressCalled    func([]byte, []byte) error
	SetOwnerAddressCalled       func([]byte)
	GetOwnerAddressCalled       func() []byte
	SetUserNameCalled           func(userName []byte)
	GetUserNameCalled           func() []byte
	IsGuardedCalled             func() bool
	GetAllLeavesCalled          func(leavesChannels *common.TrieIteratorChannels, ctx context.Context) error
}

// AddressBytes -
func (aas *StateUserAccountHandlerStub) AddressBytes() []byte {
	if aas.AddressBytesCalled != nil {
		return aas.AddressBytesCalled()
	}

	return nil
}

// IncreaseNonce -
func (aas *StateUserAccountHandlerStub) IncreaseNonce(nonce uint64) {
	if aas.IncreaseNonceCalled != nil {
		aas.IncreaseNonceCalled(nonce)
	}
}

// GetNonce -
func (aas *StateUserAccountHandlerStub) GetNonce() uint64 {
	if aas.GetNonceCalled != nil {
		return aas.GetNonceCalled()
	}
	return 0
}

// SetCode -
func (aas *StateUserAccountHandlerStub) SetCode(code []byte) {
	if aas.SetCodeCalled != nil {
		aas.SetCodeCalled(code)
	}
}

// SetCodeMetadata -
func (aas *StateUserAccountHandlerStub) SetCodeMetadata(codeMetadata []byte) {
	if aas.SetCodeMetadataCalled != nil {
		aas.SetCodeMetadataCalled(codeMetadata)
	}
}

// GetCodeMetadata -
func (aas *StateUserAccountHandlerStub) GetCodeMetadata() []byte {
	if aas.GetCodeMetadataCalled != nil {
		return aas.GetCodeMetadataCalled()
	}
	return nil
}

// SetCodeHash -
func (aas *StateUserAccountHandlerStub) SetCodeHash(codeHash []byte) {
	if aas.SetCodeHashCalled != nil {
		aas.SetCodeHashCalled(codeHash)
	}
}

// GetCodeHash -
func (aas *StateUserAccountHandlerStub) GetCodeHash() []byte {
	if aas.GetCodeHashCalled != nil {
		return aas.GetCodeHashCalled()
	}
	return nil
}

// SetRootHash -
func (aas *StateUserAccountHandlerStub) SetRootHash(rootHash []byte) {
	if aas.SetRootHashCalled != nil {
		aas.SetRootHashCalled(rootHash)
	}
}

// GetRootHash -
func (aas *StateUserAccountHandlerStub) GetRootHash() []byte {
	if aas.GetRootHashCalled != nil {
		return aas.GetRootHashCalled()
	}
	return nil
}

// SetDataTrie -
func (aas *StateUserAccountHandlerStub) SetDataTrie(trie common.Trie) {
	if aas.SetDataTrieCalled != nil {
		aas.SetDataTrieCalled(trie)
	}
}

// DataTrie -
func (aas *StateUserAccountHandlerStub) DataTrie() common.DataTrieHandler {
	if aas.DataTrieCalled != nil {
		return aas.DataTrieCalled()
	}
	return nil
}

// RetrieveValue -
func (aas *StateUserAccountHandlerStub) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if aas.RetrieveValueCalled != nil {
		return aas.RetrieveValueCalled(key)
	}
	return nil, 0, nil
}

// SaveKeyValue -
func (aas *StateUserAccountHandlerStub) SaveKeyValue(key []byte, value []byte) error {
	if aas.SaveKeyValueCalled != nil {
		return aas.SaveKeyValueCalled(key, value)
	}
	return nil
}

// AddToBalance -
func (aas *StateUserAccountHandlerStub) AddToBalance(value *big.Int) error {
	if aas.AddToBalanceCalled != nil {
		return aas.AddToBalanceCalled(value)
	}
	return nil
}

// SubFromBalance -
func (aas *StateUserAccountHandlerStub) SubFromBalance(value *big.Int) error {
	if aas.SubFromBalanceCalled != nil {
		return aas.SubFromBalanceCalled(value)
	}
	return nil
}

// GetBalance -
func (aas *StateUserAccountHandlerStub) GetBalance() *big.Int {
	if aas.GetBalanceCalled != nil {
		return aas.GetBalanceCalled()
	}
	return big.NewInt(0)
}

// ClaimDeveloperRewards -
func (aas *StateUserAccountHandlerStub) ClaimDeveloperRewards(senderAddr []byte) (*big.Int, error) {
	if aas.ClaimDeveloperRewardsCalled != nil {
		return aas.ClaimDeveloperRewardsCalled(senderAddr)
	}
	return big.NewInt(0), nil
}

// AddToDeveloperReward -
func (aas *StateUserAccountHandlerStub) AddToDeveloperReward(val *big.Int) {
	if aas.AddToDeveloperRewardCalled != nil {
		aas.AddToDeveloperRewardCalled(val)
	}
}

// GetDeveloperReward -
func (aas *StateUserAccountHandlerStub) GetDeveloperReward() *big.Int {
	if aas.GetDeveloperRewardCalled != nil {
		return aas.GetDeveloperRewardCalled()
	}
	return nil
}

// ChangeOwnerAddress -
func (aas *StateUserAccountHandlerStub) ChangeOwnerAddress(senderAddr []byte, newOwnerAddr []byte) error {
	if aas.ChangeOwnerAddressCalled != nil {
		return aas.ChangeOwnerAddressCalled(senderAddr, newOwnerAddr)
	}
	return nil
}

// SetOwnerAddress -
func (aas *StateUserAccountHandlerStub) SetOwnerAddress(address []byte) {
	if aas.SetOwnerAddressCalled != nil {
		aas.SetOwnerAddressCalled(address)
	}
}

// GetOwnerAddress -
func (aas *StateUserAccountHandlerStub) GetOwnerAddress() []byte {
	if aas.GetOwnerAddressCalled != nil {
		return aas.GetOwnerAddressCalled()
	}
	return nil
}

// SetUserName -
func (aas *StateUserAccountHandlerStub) SetUserName(userName []byte) {
	if aas.SetUserNameCalled != nil {
		aas.SetUserNameCalled(userName)
	}
}

// GetUserName -
func (aas *StateUserAccountHandlerStub) GetUserName() []byte {
	if aas.GetUserNameCalled != nil {
		return aas.GetUserNameCalled()
	}
	return nil
}

// IsGuarded -
func (aas *StateUserAccountHandlerStub) IsGuarded() bool {
	if aas.IsGuardedCalled != nil {
		return aas.IsGuardedCalled()
	}
	return false
}

// GetAllLeaves -
func (aas *StateUserAccountHandlerStub) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context) error {
	if aas.GetAllLeavesCalled != nil {
		return aas.GetAllLeavesCalled(leavesChannels, ctx)
	}

	return nil
}

// IsInterfaceNil -
func (aas *StateUserAccountHandlerStub) IsInterfaceNil() bool {
	return aas == nil
}
