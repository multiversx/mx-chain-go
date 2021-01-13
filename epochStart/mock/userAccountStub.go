package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

var _ state.UserAccountHandler = (*UserAccountStub)(nil)

// UserAccountStub -
type UserAccountStub struct {
	AddToBalanceCalled    func(value *big.Int) error
	DataTrieTrackerCalled func() state.DataTrieTracker
}

// HasNewCode -
func (u *UserAccountStub) HasNewCode() bool {
	return false
}

// SetUserName -
func (u *UserAccountStub) SetUserName(_ []byte) {
}

// GetUserName -
func (u *UserAccountStub) GetUserName() []byte {
	return nil
}

// AddToBalance -
func (u *UserAccountStub) AddToBalance(value *big.Int) error {
	if u.AddToBalanceCalled != nil {
		return u.AddToBalanceCalled(value)
	}
	return nil
}

// SubFromBalance -
func (u *UserAccountStub) SubFromBalance(_ *big.Int) error {
	return nil
}

// GetBalance -
func (u *UserAccountStub) GetBalance() *big.Int {
	return nil
}

// ClaimDeveloperRewards -
func (u *UserAccountStub) ClaimDeveloperRewards([]byte) (*big.Int, error) {
	return nil, nil
}

// AddToDeveloperReward -
func (u *UserAccountStub) AddToDeveloperReward(*big.Int) {

}

// GetDeveloperReward -
func (u *UserAccountStub) GetDeveloperReward() *big.Int {
	return nil
}

// ChangeOwnerAddress -
func (u *UserAccountStub) ChangeOwnerAddress([]byte, []byte) error {
	return nil
}

// SetOwnerAddress -
func (u *UserAccountStub) SetOwnerAddress([]byte) {

}

// GetOwnerAddress -
func (u *UserAccountStub) GetOwnerAddress() []byte {
	return nil
}

// AddressBytes -
func (u *UserAccountStub) AddressBytes() []byte {
	return nil
}

//IncreaseNonce -
func (u *UserAccountStub) IncreaseNonce(_ uint64) {
}

// GetNonce -
func (u *UserAccountStub) GetNonce() uint64 {
	return 0
}

// SetCode -
func (u *UserAccountStub) SetCode(_ []byte) {

}

// GetCode -
func (u *UserAccountStub) GetCode() []byte {
	return nil
}

// SetCodeMetadata -
func (u *UserAccountStub) SetCodeMetadata(_ []byte) {
}

// GetCodeMetadata -
func (u *UserAccountStub) GetCodeMetadata() []byte {
	return nil
}

// SetCodeHash -
func (u *UserAccountStub) SetCodeHash([]byte) {

}

// GetCodeHash -
func (u *UserAccountStub) GetCodeHash() []byte {
	return nil
}

// SetRootHash -
func (u *UserAccountStub) SetRootHash([]byte) {

}

// GetRootHash -
func (u *UserAccountStub) GetRootHash() []byte {
	return nil
}

// SetDataTrie -
func (u *UserAccountStub) SetDataTrie(_ data.Trie) {

}

// DataTrie -
func (u *UserAccountStub) DataTrie() data.Trie {
	return nil
}

// DataTrieTracker -
func (u *UserAccountStub) DataTrieTracker() state.DataTrieTracker {
	if u.DataTrieTrackerCalled != nil {
		return u.DataTrieTrackerCalled()
	}
	return nil
}

// IsInterfaceNil -
func (u *UserAccountStub) IsInterfaceNil() bool {
	return false
}
