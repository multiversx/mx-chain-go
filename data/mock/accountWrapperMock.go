package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
)

type AccountWrapMock struct {
	MockValue         int
	dataTrie          trie.Trie
	nonce             uint64
	code              []byte
	codeHash          []byte
	rootHash          []byte
	address           state.AddressContainer
	tracker           state.AccountTracker
	trackableDataTrie state.DataTrieTracker

	SetNonceWithJournalCalled    func(nonce uint64) error    `json:"-"`
	SetCodeHashWithJournalCalled func(codeHash []byte) error `json:"-"`
	SetRootHashWithJournalCalled func([]byte) error          `json:"-"`
	SetCodeWithJournalCalled     func([]byte) error          `json:"-"`
}

func NewAccountWrapMock(adr state.AddressContainer, tracker state.AccountTracker) *AccountWrapMock {
	return &AccountWrapMock{
		address:           adr,
		tracker:           tracker,
		trackableDataTrie: state.NewTrackableDataTrie(nil),
	}
}

func (awm *AccountWrapMock) IsInterfaceNil() bool {
	if awm == nil {
		return true
	}
	return false
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

func (awm *AccountWrapMock) SetCodeWithJournal(code []byte) error {
	return awm.SetCodeWithJournalCalled(code)
}

func (awm *AccountWrapMock) DataTrie() trie.Trie {
	return awm.dataTrie
}

func (awm *AccountWrapMock) SetDataTrie(trie trie.Trie) {
	awm.dataTrie = trie
	awm.trackableDataTrie.SetDataTrie(trie)
}

func (awm *AccountWrapMock) DataTrieTracker() state.DataTrieTracker {
	return awm.trackableDataTrie
}

func (awm *AccountWrapMock) SetDataTrieTracker(tracker state.DataTrieTracker) {
	awm.trackableDataTrie = tracker
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (awm *AccountWrapMock) SetNonceWithJournal(nonce uint64) error {
	return awm.SetNonceWithJournalCalled(nonce)
}

//SetNonce saves the nonce to the account
func (awm *AccountWrapMock) SetNonce(nonce uint64) {
	awm.nonce = nonce
}

// GetNonce gets the nonce of the account
func (awm *AccountWrapMock) GetNonce() uint64 {
	return awm.nonce
}
