package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// UserAccountStub -
type UserAccountStub struct {
	AddressContainerCalled           func() state.AddressContainer
	GetCodeHashCalled                func() []byte
	SetCodeHashCalled                func([]byte)
	SetCodeHashWithJournalCalled     func([]byte) error
	GetCodeCalled                    func() []byte
	SetCodeCalled                    func(code []byte)
	SetNonceCalled                   func(nonce uint64)
	GetNonceCalled                   func() uint64
	SetNonceWithJournalCalled        func(nonce uint64) error
	GetRootHashCalled                func() []byte
	DataTrieCalled                   func() data.Trie
	DataTrieTrackerCalled            func() state.DataTrieTracker
	ClaimDeveloperRewardsCalled      func(sndAddress []byte) (*big.Int, error)
	ChangeOwnerAddressCalled         func(sndAddress []byte, newAddress []byte) error
	AddToDeveloperRewardCalled       func(value *big.Int) error
	AddToBalanceCalled               func(value *big.Int) error
	SubFromBalanceCalled             func(value *big.Int) error
	GetBalanceCalled                 func() *big.Int
	SetOwnerAddressWithJournalCalled func(ownerAddress []byte) error
	GetOwnerAddressCalled            func() []byte
	SetRootHashCalled                func([]byte)
	SetDataTrieCalled                func(trie data.Trie)
}

// SubFromBalance -
func (u *UserAccountStub) SubFromBalance(value *big.Int) error {
	if u.SubFromBalanceCalled != nil {
		return u.SubFromBalanceCalled(value)
	}
	return nil
}

// AddressContainer -
func (u *UserAccountStub) AddressContainer() state.AddressContainer {
	if u.AddressContainerCalled != nil {
		return u.AddressContainerCalled()
	}
	return &AddressMock{}
}

// GetCodeHash -
func (u *UserAccountStub) GetCodeHash() []byte {
	if u.GetCodeCalled != nil {
		return u.GetCodeCalled()
	}
	return nil
}

// SetCodeHash -
func (u *UserAccountStub) SetCodeHash(code []byte) {
	if u.SetCodeCalled != nil {
		u.SetCodeCalled(code)
	}
}

// SetCodeHashWithJournal -
func (u *UserAccountStub) SetCodeHashWithJournal(codeHash []byte) error {
	if u.SetCodeHashWithJournalCalled != nil {
		return u.SetCodeHashWithJournalCalled(codeHash)
	}
	return nil
}

// GetCode -
func (u *UserAccountStub) GetCode() []byte {
	if u.GetCodeCalled != nil {
		return u.GetCodeCalled()
	}
	return nil
}

// SetCode -
func (u *UserAccountStub) SetCode(code []byte) {
	if u.SetCodeCalled != nil {
		u.SetCodeCalled(code)
	}
}

// SetNonce -
func (u *UserAccountStub) SetNonce(nonce uint64) {
	if u.SetNonceCalled != nil {
		u.SetNonceCalled(nonce)
	}
}

// GetNonce -
func (u *UserAccountStub) GetNonce() uint64 {
	if u.GetNonceCalled != nil {
		return u.GetNonceCalled()
	}
	return 0
}

// SetNonceWithJournal -
func (u *UserAccountStub) SetNonceWithJournal(nonce uint64) error {
	if u.SetNonceWithJournalCalled != nil {
		return u.SetNonceWithJournalCalled(nonce)
	}
	return nil
}

// GetRootHash -
func (u *UserAccountStub) GetRootHash() []byte {
	if u.GetRootHashCalled != nil {
		return u.GetRootHashCalled()
	}
	return nil
}

// SetRootHash -
func (u *UserAccountStub) SetRootHash(hash []byte) {
	if u.SetRootHashCalled != nil {
		u.SetRootHashCalled(hash)
	}
}

// DataTrie -
func (u *UserAccountStub) DataTrie() data.Trie {
	if u.DataTrieCalled != nil {
		return u.DataTrieCalled()
	}
	return &TrieStub{}
}

// SetDataTrie -
func (u *UserAccountStub) SetDataTrie(trie data.Trie) {
	if u.SetDataTrieCalled != nil {
		u.SetDataTrieCalled(trie)
	}
}

// DataTrieTracker -
func (u *UserAccountStub) DataTrieTracker() state.DataTrieTracker {
	if u.DataTrieCalled != nil {
		return u.DataTrieTrackerCalled()
	}
	return &DataTrieTrackerStub{}
}

// IsInterfaceNil -
func (u *UserAccountStub) IsInterfaceNil() bool {
	return u == nil
}

// ClaimDeveloperRewards -
func (u *UserAccountStub) ClaimDeveloperRewards(sndAddress []byte) (*big.Int, error) {
	if u.ClaimDeveloperRewardsCalled != nil {
		return u.ClaimDeveloperRewardsCalled(sndAddress)
	}
	return big.NewInt(0), nil
}

// ChangeOwnerAddress -
func (u *UserAccountStub) ChangeOwnerAddress(sndAddress []byte, newAddress []byte) error {
	if u.ChangeOwnerAddressCalled != nil {
		return u.ChangeOwnerAddressCalled(sndAddress, newAddress)
	}
	return nil
}

// AddToDeveloperReward -
func (u *UserAccountStub) AddToDeveloperReward(value *big.Int) error {
	if u.AddToDeveloperRewardCalled != nil {
		return u.AddToDeveloperRewardCalled(value)
	}
	return nil
}

// AddToBalance -
func (u *UserAccountStub) AddToBalance(value *big.Int) error {
	if u.AddToBalanceCalled != nil {
		return u.AddToBalanceCalled(value)
	}
	return nil
}

// GetBalance -
func (u *UserAccountStub) GetBalance() *big.Int {
	if u.GetBalanceCalled != nil {
		return u.GetBalanceCalled()
	}
	return big.NewInt(0)
}

// SetOwnerAddressWithJournal -
func (u *UserAccountStub) SetOwnerAddressWithJournal(ownerAddress []byte) error {
	if u.SetOwnerAddressWithJournalCalled != nil {
		return u.SetOwnerAddressWithJournalCalled(ownerAddress)
	}
	return nil
}

// GetOwnerAddress -
func (u *UserAccountStub) GetOwnerAddress() []byte {
	if u.GetOwnerAddressCalled != nil {
		return u.GetOwnerAddressCalled()
	}
	return nil
}
