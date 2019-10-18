package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type PeerAccountHandlerMock struct {
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
	SetRootHashWithJournalCalled func([]byte) error          `json:"-"`
	SetCodeWithJournalCalled     func(codeHash []byte) error `json:"-"`

	IncreaseLeaderSuccessRateWithJournalCalled    func() error
	DecreaseLeaderSuccessRateWithJournalCalled    func() error
	IncreaseValidatorSuccessRateWithJournalCalled func() error
	DecreaseValidatorSuccessRateWithJournalCalled func() error
}

func (awm *PeerAccountHandlerMock) GetCodeHash() []byte {
	return awm.codeHash
}

func (awm *PeerAccountHandlerMock) SetCodeHash(codeHash []byte) {
	awm.codeHash = codeHash
}

func (awm *PeerAccountHandlerMock) SetCodeHashWithJournal(codeHash []byte) error {
	return awm.SetCodeHashWithJournalCalled(codeHash)
}

func (awm *PeerAccountHandlerMock) GetCode() []byte {
	return awm.code
}

func (awm *PeerAccountHandlerMock) GetRootHash() []byte {
	return awm.rootHash
}

func (awm *PeerAccountHandlerMock) SetRootHash(rootHash []byte) {
	awm.rootHash = rootHash
}

func (awm *PeerAccountHandlerMock) SetRootHashWithJournal(rootHash []byte) error {
	return awm.SetRootHashWithJournalCalled(rootHash)
}

func (awm *PeerAccountHandlerMock) SetNonceWithJournal(nonce uint64) error {
	return awm.SetNonceWithJournalCalled(nonce)
}

func (awm *PeerAccountHandlerMock) AddressContainer() state.AddressContainer {
	return awm.address
}

func (awm *PeerAccountHandlerMock) SetCode(code []byte) {
	awm.code = code
}

func (awm *PeerAccountHandlerMock) DataTrie() data.Trie {
	return awm.dataTrie
}

func (awm *PeerAccountHandlerMock) SetDataTrie(trie data.Trie) {
	awm.dataTrie = trie
	awm.trackableDataTrie.SetDataTrie(trie)
}

func (awm *PeerAccountHandlerMock) DataTrieTracker() state.DataTrieTracker {
	return awm.trackableDataTrie
}

func (awm *PeerAccountHandlerMock) SetDataTrieTracker(tracker state.DataTrieTracker) {
	awm.trackableDataTrie = tracker
}

func (awm *PeerAccountHandlerMock) SetNonce(nonce uint64) {
	awm.nonce = nonce
}

func (awm *PeerAccountHandlerMock) GetNonce() uint64 {
	return awm.nonce
}

func (awm *PeerAccountHandlerMock) IncreaseLeaderSuccessRateWithJournal() error {
	if awm.IncreaseLeaderSuccessRateWithJournalCalled != nil {
		return awm.IncreaseLeaderSuccessRateWithJournalCalled()
	}
	return nil
}

func (awm *PeerAccountHandlerMock) DecreaseLeaderSuccessRateWithJournal() error {
	if awm.DecreaseLeaderSuccessRateWithJournalCalled != nil {
		return awm.DecreaseLeaderSuccessRateWithJournalCalled()
	}
	return nil
}

func (awm *PeerAccountHandlerMock) IncreaseValidatorSuccessRateWithJournal() error {
	if awm.IncreaseValidatorSuccessRateWithJournalCalled != nil {
		return awm.IncreaseValidatorSuccessRateWithJournalCalled()
	}
	return nil
}

func (awm *PeerAccountHandlerMock) DecreaseValidatorSuccessRateWithJournal() error {
	if awm.DecreaseValidatorSuccessRateWithJournalCalled != nil {
		return awm.DecreaseValidatorSuccessRateWithJournalCalled()
	}
	return nil
}

func (awm *PeerAccountHandlerMock) IsInterfaceNil() bool {
	if awm == nil {
		return true
	}
	return false
}
