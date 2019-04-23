package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type AccountWrapMock struct {
	dataTrie trie.PatriciaMerkelTree
	code     []byte
	codeHash []byte
	rootHash []byte
	Address  state.AddressContainer
	Tracker  state.AccountTracker

	SetCodeHashWithJournalCalled func(codeHash []byte) error
	SetRootHashWithJournalCalled func([]byte) error
}

func (awm *AccountWrapMock) DataTrieTracker() state.DataTrieTracker {
	panic("implement me")
}

func NewAccountWrapMock(adr state.AddressContainer, tracker state.AccountTracker) *AccountWrapMock {
	return &AccountWrapMock{
		Address: adr,
		Tracker: tracker,
	}
}

func (awm *AccountWrapMock) GetCodeHash() []byte {
	return awm.codeHash
}

func (awm *AccountWrapMock) SetCodeHash(codeHash []byte) {
	awm.codeHash = codeHash
}

func (awm *AccountWrapMock) SetCodeHashWithJournal(codeHash []byte) error {
	return awm.SetCodeHashWithJournalCalled(codeHash)
}

func (awm *AccountWrapMock) GetCode() []byte {
	return awm.code
}

func (awm *AccountWrapMock) GetRootHash() []byte {
	return awm.rootHash
}

func (awm *AccountWrapMock) SetRootHash(rootHash []byte) {
	awm.rootHash = rootHash
}

func (awm *AccountWrapMock) SetRootHashWithJournal(rootHash []byte) error {
	return awm.SetRootHashWithJournalCalled(rootHash)
}

func (awm *AccountWrapMock) AddressContainer() state.AddressContainer {
	return awm.Address
}

func (awm *AccountWrapMock) SetCode(code []byte) {
	awm.code = code
}

func (awm *AccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	return awm.dataTrie
}

func (awm *AccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	awm.dataTrie = trie
}
