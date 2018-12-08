package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type AccountWrapMock struct {
	dataTrie trie.PatriciaMerkelTree
}

func NewAccountWrapMock() *AccountWrapMock {
	return &AccountWrapMock{}
}

func (awm *AccountWrapMock) BaseAccount() *state.Account {
	panic("implement me")
}

func (awm *AccountWrapMock) AddressContainer() state.AddressContainer {
	panic("implement me")
}

func (awm *AccountWrapMock) Code() []byte {
	panic("implement me")
}

func (awm *AccountWrapMock) SetCode(code []byte) {
	panic("implement me")
}

func (awm *AccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	return awm.dataTrie
}

func (awm *AccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	awm.dataTrie = trie
}
