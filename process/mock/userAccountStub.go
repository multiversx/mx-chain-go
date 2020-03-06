package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type UserAccountStub struct {
	AddToBalanceCalled func(value *big.Int) error
}

func (u *UserAccountStub) AddToBalance(value *big.Int) error {
	if u.AddToBalanceCalled != nil {
		return u.AddToBalanceCalled(value)
	}
	return nil
}

func (u *UserAccountStub) SubFromBalance(value *big.Int) error {
	return nil
}

func (u *UserAccountStub) GetBalance() *big.Int {
	return nil
}

func (u *UserAccountStub) ClaimDeveloperRewards([]byte) (*big.Int, error) {
	return nil, nil
}

func (u *UserAccountStub) AddToDeveloperReward(*big.Int) {

}

func (u *UserAccountStub) GetDeveloperReward() *big.Int {
	return nil
}

func (u *UserAccountStub) ChangeOwnerAddress([]byte, []byte) error {
	return nil
}

func (u *UserAccountStub) SetOwnerAddress([]byte) {

}

func (u *UserAccountStub) GetOwnerAddress() []byte {
	return nil
}

func (u *UserAccountStub) AddressContainer() state.AddressContainer {
	return nil
}

func (u *UserAccountStub) SetNonce(nonce uint64) {

}

func (u *UserAccountStub) GetNonce() uint64 {
	return 0
}

func (u *UserAccountStub) SetCode(code []byte) {

}

func (u *UserAccountStub) GetCode() []byte {
	return nil
}

func (u *UserAccountStub) SetCodeHash([]byte) {

}

func (u *UserAccountStub) GetCodeHash() []byte {
	return nil
}

func (u *UserAccountStub) SetRootHash([]byte) {

}

func (u *UserAccountStub) GetRootHash() []byte {
	return nil
}

func (u *UserAccountStub) SetDataTrie(trie data.Trie) {

}

func (u *UserAccountStub) DataTrie() data.Trie {
	return nil
}

func (u *UserAccountStub) DataTrieTracker() state.DataTrieTracker {
	return nil
}

func (u *UserAccountStub) IsInterfaceNil() bool {
	return false
}
