package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// PeerAccountHandlerMock -
type PeerAccountHandlerMock struct {
	MockValue         int
	dataTrie          data.Trie
	nonce             uint64
	code              []byte
	codeHash          []byte
	rootHash          []byte
	address           state.AddressContainer
	trackableDataTrie state.DataTrieTracker

	SetNonceWithJournalCalled      func(nonce uint64) error
	SetCodeHashWithJournalCalled   func(codeHash []byte) error
	SetRootHashWithJournalCalled   func([]byte) error
	RatingCalled                   func() uint32
	SetCodeWithJournalCalled       func(codeHash []byte) error
	SetRatingWithJournalCalled     func(rating uint32) error
	GetTempRatingCalled            func() uint32
	SetTempRatingWithJournalCalled func(rating uint32) error

	IncreaseLeaderSuccessRateWithJournalCalled    func(value uint32) error
	DecreaseLeaderSuccessRateWithJournalCalled    func(value uint32) error
	IncreaseValidatorSuccessRateWithJournalCalled func(value uint32) error
	DecreaseValidatorSuccessRateWithJournalCalled func(value uint32) error
	AddToAccumulatedFeesCalled                    func(value *big.Int) error
}

// AddToAccumulatedFees -
func (pahm *PeerAccountHandlerMock) AddToAccumulatedFees(value *big.Int) error {
	if pahm.AddToAccumulatedFeesCalled != nil {
		return pahm.AddToAccumulatedFeesCalled(value)
	}
	return nil
}

// GetCodeHash -
func (pahm *PeerAccountHandlerMock) GetCodeHash() []byte {
	return pahm.codeHash
}

// SetCodeHash -
func (pahm *PeerAccountHandlerMock) SetCodeHash(codeHash []byte) {
	pahm.codeHash = codeHash
}

// SetCodeHashWithJournal -
func (pahm *PeerAccountHandlerMock) SetCodeHashWithJournal(codeHash []byte) error {
	return pahm.SetCodeHashWithJournalCalled(codeHash)
}

// GetCode -
func (pahm *PeerAccountHandlerMock) GetCode() []byte {
	return pahm.code
}

// GetRootHash -
func (pahm *PeerAccountHandlerMock) GetRootHash() []byte {
	return pahm.rootHash
}

// SetRootHash -
func (pahm *PeerAccountHandlerMock) SetRootHash(rootHash []byte) {
	pahm.rootHash = rootHash
}

// SetNonceWithJournal -
func (pahm *PeerAccountHandlerMock) SetNonceWithJournal(nonce uint64) error {
	return pahm.SetNonceWithJournalCalled(nonce)
}

// AddressContainer -
func (pahm *PeerAccountHandlerMock) AddressContainer() state.AddressContainer {
	return pahm.address
}

// SetCode -
func (pahm *PeerAccountHandlerMock) SetCode(code []byte) {
	pahm.code = code
}

// DataTrie -
func (pahm *PeerAccountHandlerMock) DataTrie() data.Trie {
	return pahm.dataTrie
}

// SetDataTrie -
func (pahm *PeerAccountHandlerMock) SetDataTrie(trie data.Trie) {
	pahm.dataTrie = trie
	pahm.trackableDataTrie.SetDataTrie(trie)
}

// DataTrieTracker -
func (pahm *PeerAccountHandlerMock) DataTrieTracker() state.DataTrieTracker {
	return pahm.trackableDataTrie
}

// SetDataTrieTracker -
func (pahm *PeerAccountHandlerMock) SetDataTrieTracker(tracker state.DataTrieTracker) {
	pahm.trackableDataTrie = tracker
}

// SetNonce -
func (pahm *PeerAccountHandlerMock) SetNonce(nonce uint64) {
	pahm.nonce = nonce
}

// GetNonce -
func (pahm *PeerAccountHandlerMock) GetNonce() uint64 {
	return pahm.nonce
}

// IncreaseLeaderSuccessRateWithJournal -
func (pahm *PeerAccountHandlerMock) IncreaseLeaderSuccessRateWithJournal(value uint32) error {
	if pahm.IncreaseLeaderSuccessRateWithJournalCalled != nil {
		return pahm.IncreaseLeaderSuccessRateWithJournalCalled(value)
	}
	return nil
}

// DecreaseLeaderSuccessRateWithJournal -
func (pahm *PeerAccountHandlerMock) DecreaseLeaderSuccessRateWithJournal(value uint32) error {
	if pahm.DecreaseLeaderSuccessRateWithJournalCalled != nil {
		return pahm.DecreaseLeaderSuccessRateWithJournalCalled(value)
	}
	return nil
}

// IncreaseValidatorSuccessRateWithJournal -
func (pahm *PeerAccountHandlerMock) IncreaseValidatorSuccessRateWithJournal(value uint32) error {
	if pahm.IncreaseValidatorSuccessRateWithJournalCalled != nil {
		return pahm.IncreaseValidatorSuccessRateWithJournalCalled(value)
	}
	return nil
}

// DecreaseValidatorSuccessRateWithJournal -
func (pahm *PeerAccountHandlerMock) DecreaseValidatorSuccessRateWithJournal(value uint32) error {
	if pahm.DecreaseValidatorSuccessRateWithJournalCalled != nil {
		return pahm.DecreaseValidatorSuccessRateWithJournalCalled(value)
	}
	return nil
}

// GetRating -
func (pahm *PeerAccountHandlerMock) GetRating() uint32 {
	if pahm.SetRatingWithJournalCalled != nil {
		return pahm.RatingCalled()
	}
	return 10
}

// SetRatingWithJournal -
func (pahm *PeerAccountHandlerMock) SetRatingWithJournal(rating uint32) error {
	if pahm.SetRatingWithJournalCalled != nil {
		return pahm.SetRatingWithJournalCalled(rating)
	}
	return nil
}

// GetTempRating -
func (pahm *PeerAccountHandlerMock) GetTempRating() uint32 {
	if pahm.GetTempRatingCalled != nil {
		return pahm.GetTempRatingCalled()
	}
	return 10
}

// SetTempRatingWithJournal -
func (pahm *PeerAccountHandlerMock) SetTempRatingWithJournal(rating uint32) error {
	if pahm.SetTempRatingWithJournalCalled != nil {
		return pahm.SetTempRatingWithJournalCalled(rating)
	}
	return nil
}

// IsInterfaceNil -
func (pahm *PeerAccountHandlerMock) IsInterfaceNil() bool {
	if pahm == nil {
		return true
	}
	return false
}
