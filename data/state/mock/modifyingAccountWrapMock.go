package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type ModifyingAccountWrapMock struct {
	nonce    uint64
	balance  big.Int
	codeHash []byte
	rootHash []byte
}

func NewModifyingAccountWrapMock() *ModifyingAccountWrapMock {
	return &ModifyingAccountWrapMock{}
}

func (mawm *ModifyingAccountWrapMock) Nonce() uint64 {
	return mawm.nonce
}

func (mawm *ModifyingAccountWrapMock) SetNonce(nonce uint64) {
	mawm.nonce = nonce
}

func (mawm *ModifyingAccountWrapMock) CodeHash() []byte {
	return mawm.codeHash
}

func (mawm *ModifyingAccountWrapMock) SetCodeHash(codeHash []byte) {
	mawm.codeHash = codeHash
}

func (mawm *ModifyingAccountWrapMock) RootHash() []byte {
	return mawm.rootHash
}

func (mawm *ModifyingAccountWrapMock) SetRootHash(rootHash []byte) {
	mawm.rootHash = rootHash
}

func (mawm *ModifyingAccountWrapMock) Balance() big.Int {
	return mawm.balance
}

func (mawm *ModifyingAccountWrapMock) SetBalance(balance big.Int) {
	mawm.balance = balance
}

func (mawm *ModifyingAccountWrapMock) LoadFromDbAccount(source state.DbAccountContainer) error {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) SaveToDbAccount(target state.DbAccountContainer) error {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) AddressContainer() state.AddressContainer {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) Code() []byte {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) SetCode(code []byte) {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) ClearDataCaches() {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) DirtyData() map[string][]byte {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) OriginalValue(key []byte) []byte {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) RetrieveValue(key []byte) ([]byte, error) {
	panic("implement me")
}

func (mawm *ModifyingAccountWrapMock) SaveKeyValue(key []byte, value []byte) {
	panic("implement me")
}
