package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type AccountWrapMock struct {
	MockValue         int
	dataTrie          trie.PatriciaMerkelTree
	code              []byte
	codeHash          []byte
	rootHash          []byte
	address           state.AddressContainer
	tracker           state.AccountTracker
	trackableDataTrie state.DataTrieTracker

	SetCodeHashWithJournalCalled func(codeHash []byte) error `json:"-"`
	SetRootHashWithJournalCalled func([]byte) error          `json:"-"`
}

func NewAccountWrapMock(adr state.AddressContainer, tracker state.AccountTracker) *AccountWrapMock {
	return &AccountWrapMock{
		address:           adr,
		tracker:           tracker,
		trackableDataTrie: state.NewTrackableDataTrie(nil),
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
	return awm.address
}

func (awm *AccountWrapMock) SetCode(code []byte) {
	awm.code = code
}

func (awm *AccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	return awm.dataTrie
}

func (awm *AccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	awm.dataTrie = trie
	awm.trackableDataTrie.SetDataTrie(trie)
}

func (awm *AccountWrapMock) DataTrieTracker() state.DataTrieTracker {
	return awm.trackableDataTrie
}

func (awm *AccountWrapMock) SetDataTrieTracker(tracker state.DataTrieTracker) {
	awm.trackableDataTrie = tracker
}

func (awm *AccountWrapMock) IsInterfaceNil() bool {
	if awm == nil {
		return true
	}

	return false
}
