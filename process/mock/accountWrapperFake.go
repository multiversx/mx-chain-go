package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AccountWrapMock -
type AccountWrapMock struct {
	MockValue         int
	dataTrie          data.Trie
	nonce             uint64
	rating            uint32
	tempRating        uint32
	consecutiveMisses uint32
	code              []byte
	codeHash          []byte
	rootHash          []byte
	address           state.AddressContainer
	tracker           state.AccountTracker
	trackableDataTrie state.DataTrieTracker

	SetNonceWithJournalCalled    func(nonce uint64) error    `json:"-"`
	SetCodeHashWithJournalCalled func(codeHash []byte) error `json:"-"`
	SetCodeWithJournalCalled     func(codeHash []byte) error `json:"-"`
}

// NewAccountWrapMock -
func NewAccountWrapMock(adr state.AddressContainer, tracker state.AccountTracker) *AccountWrapMock {
	return &AccountWrapMock{
		address:           adr,
		tracker:           tracker,
		trackableDataTrie: state.NewTrackableDataTrie([]byte("identifier"), nil),
	}
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

// SetNonceWithJournal -
func (awm *AccountWrapMock) SetNonceWithJournal(nonce uint64) error {
	return awm.SetNonceWithJournalCalled(nonce)
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

// SetNonce -
func (awm *AccountWrapMock) SetNonce(nonce uint64) {
	awm.nonce = nonce
}

// GetNonce -
func (awm *AccountWrapMock) GetNonce() uint64 {
	return awm.nonce
}

// GetConsecutiveProposerMisses -
func (awm *AccountWrapMock) GetConsecutiveProposerMisses() uint32 {
	return awm.consecutiveMisses
}

// SetConsecutiveProposerMissesWithJournal -
func (awm *AccountWrapMock) SetConsecutiveProposerMissesWithJournal(consecutiveMisses uint32) error {
	awm.consecutiveMisses = consecutiveMisses
	return nil
}

// IsInterfaceNil -
func (awm *AccountWrapMock) IsInterfaceNil() bool {
	return awm == nil
}

func (awm *AccountWrapMock) AddToAccumulatedFees(value *big.Int) error {
	return nil
}
func (awm *AccountWrapMock) IncreaseLeaderSuccessRateWithJournal(value uint32) error {
	return nil
}
func (awm *AccountWrapMock) DecreaseLeaderSuccessRateWithJournal(value uint32) error {
	return nil
}
func (awm *AccountWrapMock) IncreaseValidatorSuccessRateWithJournal(value uint32) error {
	return nil
}
func (awm *AccountWrapMock) DecreaseValidatorSuccessRateWithJournal(value uint32) error {
	return nil
}
func (awm *AccountWrapMock) IncreaseNumSelectedInSuccessBlocks() error {
	return nil
}
func (awm *AccountWrapMock) GetRating() uint32 {
	return awm.rating
}
func (awm *AccountWrapMock) SetRatingWithJournal(rating uint32) error {
	awm.rating = rating
	return nil
}
func (awm *AccountWrapMock) GetTempRating() uint32 {
	return awm.tempRating
}
func (awm *AccountWrapMock) SetTempRatingWithJournal(tempRating uint32) error {
	awm.tempRating = tempRating
	return nil
}

func (awm *AccountWrapMock) ResetAtNewEpoch() error {
	return nil
}
func (awm *AccountWrapMock) SetRewardAddressWithJournal(address []byte) error {
	return nil
}
func (awm *AccountWrapMock) SetSchnorrPublicKeyWithJournal(address []byte) error {
	return nil
}
func (awm *AccountWrapMock) SetBLSPublicKeyWithJournal(address []byte) error {
	return nil
}
func (awm *AccountWrapMock) SetStakeWithJournal(stake *big.Int) error {
	return nil
}
