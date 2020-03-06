package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AccountWrapMock -
type AccountWrapMock struct {
	MockValue         int
	dataTrie          data.Trie
	nonce             uint64
	code              []byte
	codeHash          []byte
	rootHash          []byte
	address           state.AddressContainer
	tracker           state.AccountTracker
	trackableDataTrie state.DataTrieTracker

	SetNonceWithJournalCalled    func(nonce uint64) error    `json:"-"`
	SetCodeHashWithJournalCalled func(codeHash []byte) error `json:"-"`
	SetCodeWithJournalCalled     func([]byte) error          `json:"-"`
}

// NewAccountWrapMock -
func NewAccountWrapMock(adr state.AddressContainer, tracker state.AccountTracker) *AccountWrapMock {
	return &AccountWrapMock{
		address:           adr,
		tracker:           tracker,
		trackableDataTrie: state.NewTrackableDataTrie([]byte("identifier"), nil),
	}
}

// IsInterfaceNil -
func (awm *AccountWrapMock) IsInterfaceNil() bool {
	return awm == nil
}

// GetCodeHash -
func (awm *AccountWrapMock) GetCodeHash() []byte {
	return awm.codeHash
}

// SetCodeHash -
func (awm *AccountWrapMock) SetCodeHash(codeHash []byte) {
	awm.codeHash = codeHash
}

// SetCodeHashWithJournal -
func (awm *AccountWrapMock) SetCodeHashWithJournal(codeHash []byte) error {
	return awm.SetCodeHashWithJournalCalled(codeHash)
}

// GetCode -
func (awm *AccountWrapMock) GetCode() []byte {
	return awm.code
}

// GetRootHash -
func (awm *AccountWrapMock) GetRootHash() []byte {
	return awm.rootHash
}

// SetRootHash -
func (awm *AccountWrapMock) SetRootHash(rootHash []byte) {
	awm.rootHash = rootHash
}

// AddressContainer -
func (awm *AccountWrapMock) AddressContainer() state.AddressContainer {
	return awm.address
}

// SetCode -
func (awm *AccountWrapMock) SetCode(code []byte) {
	awm.code = code
}

// DataTrie -
func (awm *AccountWrapMock) DataTrie() data.Trie {
	return awm.dataTrie
}

// SetDataTrie -
func (awm *AccountWrapMock) SetDataTrie(trie data.Trie) {
	awm.dataTrie = trie
	awm.trackableDataTrie.SetDataTrie(trie)
}

// DataTrieTracker -
func (awm *AccountWrapMock) DataTrieTracker() state.DataTrieTracker {
	return awm.trackableDataTrie
}

// SetDataTrieTracker -
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
