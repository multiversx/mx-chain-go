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

func (pahm *PeerAccountHandlerMock) GetCodeHash() []byte {
	return pahm.codeHash
}

func (pahm *PeerAccountHandlerMock) SetCodeHash(codeHash []byte) {
	pahm.codeHash = codeHash
}

func (pahm *PeerAccountHandlerMock) SetCodeHashWithJournal(codeHash []byte) error {
	return pahm.SetCodeHashWithJournalCalled(codeHash)
}

func (pahm *PeerAccountHandlerMock) GetCode() []byte {
	return pahm.code
}

func (pahm *PeerAccountHandlerMock) GetRootHash() []byte {
	return pahm.rootHash
}

func (pahm *PeerAccountHandlerMock) SetRootHash(rootHash []byte) {
	pahm.rootHash = rootHash
}

func (pahm *PeerAccountHandlerMock) SetRootHashWithJournal(rootHash []byte) error {
	return pahm.SetRootHashWithJournalCalled(rootHash)
}

func (pahm *PeerAccountHandlerMock) SetNonceWithJournal(nonce uint64) error {
	return pahm.SetNonceWithJournalCalled(nonce)
}

func (pahm *PeerAccountHandlerMock) AddressContainer() state.AddressContainer {
	return pahm.address
}

func (pahm *PeerAccountHandlerMock) SetCode(code []byte) {
	pahm.code = code
}

func (pahm *PeerAccountHandlerMock) DataTrie() data.Trie {
	return pahm.dataTrie
}

func (pahm *PeerAccountHandlerMock) SetDataTrie(trie data.Trie) {
	pahm.dataTrie = trie
	pahm.trackableDataTrie.SetDataTrie(trie)
}

func (pahm *PeerAccountHandlerMock) DataTrieTracker() state.DataTrieTracker {
	return pahm.trackableDataTrie
}

func (pahm *PeerAccountHandlerMock) SetDataTrieTracker(tracker state.DataTrieTracker) {
	pahm.trackableDataTrie = tracker
}

func (pahm *PeerAccountHandlerMock) SetNonce(nonce uint64) {
	pahm.nonce = nonce
}

func (pahm *PeerAccountHandlerMock) GetNonce() uint64 {
	return pahm.nonce
}

func (pahm *PeerAccountHandlerMock) IncreaseLeaderSuccessRateWithJournal() error {
	if pahm.IncreaseLeaderSuccessRateWithJournalCalled != nil {
		return pahm.IncreaseLeaderSuccessRateWithJournalCalled()
	}
	return nil
}

func (pahm *PeerAccountHandlerMock) DecreaseLeaderSuccessRateWithJournal() error {
	if pahm.DecreaseLeaderSuccessRateWithJournalCalled != nil {
		return pahm.DecreaseLeaderSuccessRateWithJournalCalled()
	}
	return nil
}

func (pahm *PeerAccountHandlerMock) IncreaseValidatorSuccessRateWithJournal() error {
	if pahm.IncreaseValidatorSuccessRateWithJournalCalled != nil {
		return pahm.IncreaseValidatorSuccessRateWithJournalCalled()
	}
	return nil
}

func (pahm *PeerAccountHandlerMock) DecreaseValidatorSuccessRateWithJournal() error {
	if pahm.DecreaseValidatorSuccessRateWithJournalCalled != nil {
		return pahm.DecreaseValidatorSuccessRateWithJournalCalled()
	}
	return nil
}

func (pahm *PeerAccountHandlerMock) IsInterfaceNil() bool {
	if pahm == nil {
		return true
	}
	return false
}
