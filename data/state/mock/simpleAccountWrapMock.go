package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type SimpleAccountWrapMock struct {
	dataTrie trie.PatriciaMerkelTree
}

func NewSimpleAccountWrapMock() *SimpleAccountWrapMock {
	return &SimpleAccountWrapMock{}
}

func (sawm *SimpleAccountWrapMock) Nonce() uint64 {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) SetNonce(uint64) {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) CodeHash() []byte {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) SetCodeHash([]byte) {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) RootHash() []byte {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) SetRootHash([]byte) {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) Balance() big.Int {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) SetBalance(big.Int) {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) LoadFromDbAccount(source state.DbAccountContainer) error {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) SaveToDbAccount(target state.DbAccountContainer) error {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) AddressContainer() state.AddressContainer {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) Code() []byte {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) SetCode(code []byte) {
	panic("implement me")
}

func (sawm *SimpleAccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	return sawm.dataTrie
}

func (sawm *SimpleAccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	sawm.dataTrie = trie
}
