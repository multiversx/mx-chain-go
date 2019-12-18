package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// TimeStamp is a moment defined by epoch and round
type TimeStamp struct {
	Epoch uint64
	Round uint64
}

// TimePeriod holds start and end time
type TimePeriod struct {
	StartTime TimeStamp
	EndTime   TimeStamp
}

// SignRate is used to keep the number of success and failed signings
type SignRate struct {
	NrSuccess uint32
	NrFailure uint32
}

// ValidatorApiResponse represents the data which is fetched from each validator for returning it in API call
type ValidatorApiResponse struct {
	NrLeaderSuccess    uint32 `json:"nrLeaderSuccess"`
	NrLeaderFailure    uint32 `json:"nrLeaderFailure"`
	NrValidatorSuccess uint32 `json:"nrValidatorSuccess"`
	NrValidatorFailure uint32 `json:"nrValidatorFailure"`
}

// PeerAccount is the struct used in serialization/deserialization
type PeerAccount struct {
	BLSPublicKey     []byte
	SchnorrPublicKey []byte
	Address          []byte
	Stake            *big.Int

	JailTime      TimePeriod
	PastJailTimes []TimePeriod

	CurrentShardId    uint32
	NextShardId       uint32
	NodeInWaitingList bool
	UnStakedNonce     uint64

	ValidatorSuccessRate SignRate
	LeaderSuccessRate    SignRate

	CodeHash []byte

	Rating     uint32
	TempRating uint32
	RootHash   []byte
	Nonce      uint64

	addressContainer AddressContainer
	code             []byte
	accountTracker   AccountTracker
	dataTrieTracker  DataTrieTracker
}

// NewPeerAccount creates new simple account wrapper for an PeerAccountContainer (that has just been initialized)
func NewPeerAccount(
	addressContainer AddressContainer,
	tracker AccountTracker,
) (*PeerAccount, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}
	if tracker == nil {
		return nil, ErrNilAccountTracker
	}

	return &PeerAccount{
		Stake:            big.NewInt(0),
		addressContainer: addressContainer,
		accountTracker:   tracker,
		dataTrieTracker:  NewTrackableDataTrie(nil),
	}, nil
}

// IsInterfaceNil return if there is no value under the interface
func (a *PeerAccount) IsInterfaceNil() bool {
	if a == nil {
		return true
	}
	return false
}

// AddressContainer returns the address associated with the account
func (a *PeerAccount) AddressContainer() AddressContainer {
	return a.addressContainer
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (a *PeerAccount) SetNonceWithJournal(nonce uint64) error {
	entry, err := NewBaseJournalEntryNonce(a, a.Nonce)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Nonce = nonce

	return a.accountTracker.SaveAccount(a)
}

//SetNonce saves the nonce to the account
func (a *PeerAccount) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (a *PeerAccount) GetNonce() uint64 {
	return a.Nonce
}

// GetCodeHash returns the code hash associated with this account
func (a *PeerAccount) GetCodeHash() []byte {
	return a.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (a *PeerAccount) SetCodeHash(codeHash []byte) {
	a.CodeHash = codeHash
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (a *PeerAccount) SetCodeHashWithJournal(codeHash []byte) error {
	entry, err := NewBaseJournalEntryCodeHash(a, a.CodeHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.CodeHash = codeHash

	return a.accountTracker.SaveAccount(a)
}

// GetCode gets the actual code that needs to be run in the VM
func (a *PeerAccount) GetCode() []byte {
	return a.code
}

// SetCode sets the actual code that needs to be run in the VM
func (a *PeerAccount) SetCode(code []byte) {
	a.code = code
}

// GetRootHash returns the root hash associated with this account
func (a *PeerAccount) GetRootHash() []byte {
	return a.RootHash
}

// SetRootHash sets the root hash associated with the account
func (a *PeerAccount) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// DataTrie returns the trie that holds the current account's data
func (a *PeerAccount) DataTrie() data.Trie {
	return a.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (a *PeerAccount) SetDataTrie(trie data.Trie) {
	a.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (a *PeerAccount) DataTrieTracker() DataTrieTracker {
	return a.dataTrieTracker
}

// SetAddressWithJournal sets the account's address, saving the old address before changing
func (a *PeerAccount) SetAddressWithJournal(address []byte) error {
	if len(address) < 1 {
		return ErrEmptyAddress
	}

	entry, err := NewPeerJournalEntryAddress(a, a.Address)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Address = address

	return a.accountTracker.SaveAccount(a)
}

// SetSchnorrPublicKeyWithJournal sets the account's public key, saving the old key before changing
func (a *PeerAccount) SetSchnorrPublicKeyWithJournal(pubKey []byte) error {
	if len(pubKey) < 1 {
		return ErrNilSchnorrPublicKey
	}

	entry, err := NewPeerJournalEntrySchnorrPublicKey(a, a.SchnorrPublicKey)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.SchnorrPublicKey = pubKey

	return a.accountTracker.SaveAccount(a)
}

// SetBLSPublicKeyWithJournal sets the account's bls public key, saving the old key before changing
func (a *PeerAccount) SetBLSPublicKeyWithJournal(pubKey []byte) error {
	if len(pubKey) < 1 {
		return ErrNilBLSPublicKey
	}

	entry, err := NewPeerJournalEntryBLSPublicKey(a, a.BLSPublicKey)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.BLSPublicKey = pubKey

	return a.accountTracker.SaveAccount(a)
}

// SetStakeWithJournal sets the account's stake, saving the old stake before changing
func (a *PeerAccount) SetStakeWithJournal(stake *big.Int) error {
	if stake == nil {
		return ErrNilStake
	}

	entry, err := NewPeerJournalEntryStake(a, a.Stake)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Stake = stake

	return a.accountTracker.SaveAccount(a)
}

// SetJailTimeWithJournal sets the account's jail time, saving the old state before changing
func (a *PeerAccount) SetJailTimeWithJournal(jailTime TimePeriod) error {
	entry, err := NewPeerJournalEntryJailTime(a, a.JailTime)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.JailTime = jailTime

	return a.accountTracker.SaveAccount(a)
}

// SetUnStakedNonceWithJournal sets the account's shard id, saving the old state before changing
func (a *PeerAccount) SetUnStakedNonceWithJournal(nonce uint64) error {
	entry, err := NewPeerJournalEntryUnStakedNonce(a, a.UnStakedNonce)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.UnStakedNonce = nonce

	return a.accountTracker.SaveAccount(a)
}

// SetCurrentShardIdWithJournal sets the account's shard id, saving the old state before changing
func (a *PeerAccount) SetCurrentShardIdWithJournal(shId uint32) error {
	entry, err := NewPeerJournalEntryCurrentShardId(a, a.CurrentShardId)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.CurrentShardId = shId

	return a.accountTracker.SaveAccount(a)
}

// SetNextShardIdWithJournal sets the account's shard id, saving the old state before changing
func (a *PeerAccount) SetNextShardIdWithJournal(shId uint32) error {
	entry, err := NewPeerJournalEntryNextShardId(a, a.NextShardId)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.NextShardId = shId

	return a.accountTracker.SaveAccount(a)
}

// SetNodeInWaitingListWithJournal sets the account's nodes status whether in waiting list, saving the old state before
func (a *PeerAccount) SetNodeInWaitingListWithJournal(nodeInWaitingList bool) error {
	entry, err := NewPeerJournalEntryInWaitingList(a, a.NodeInWaitingList)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.NodeInWaitingList = nodeInWaitingList

	return a.accountTracker.SaveAccount(a)
}

// IncreaseValidatorSuccessRateWithJournal increases the account's number of successful signing,
// saving the old state before changing
func (a *PeerAccount) IncreaseValidatorSuccessRateWithJournal() error {
	entry, err := NewPeerJournalEntryValidatorSuccessRate(a, a.ValidatorSuccessRate)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.ValidatorSuccessRate.NrSuccess++

	return a.accountTracker.SaveAccount(a)
}

// DecreaseValidatorSuccessRateWithJournal increases the account's number of missed signing,
// saving the old state before changing
func (a *PeerAccount) DecreaseValidatorSuccessRateWithJournal() error {
	entry, err := NewPeerJournalEntryValidatorSuccessRate(a, a.ValidatorSuccessRate)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.ValidatorSuccessRate.NrFailure++

	return a.accountTracker.SaveAccount(a)
}

// IncreaseLeaderSuccessRateWithJournal increases the account's number of successful signing,
// saving the old state before changing
func (a *PeerAccount) IncreaseLeaderSuccessRateWithJournal() error {
	entry, err := NewPeerJournalEntryLeaderSuccessRate(a, a.LeaderSuccessRate)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.LeaderSuccessRate.NrSuccess++

	return a.accountTracker.SaveAccount(a)
}

// DecreaseLeaderSuccessRateWithJournal increases the account's number of missing signing,
// saving the old state before changing
func (a *PeerAccount) DecreaseLeaderSuccessRateWithJournal() error {
	entry, err := NewPeerJournalEntryLeaderSuccessRate(a, a.LeaderSuccessRate)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.LeaderSuccessRate.NrFailure++

	return a.accountTracker.SaveAccount(a)
}

// GetRating gets the rating
func (a *PeerAccount) GetRating() uint32 {
	return a.Rating
}

// SetRatingWithJournal sets the account's rating id, saving the old state before changing
func (a *PeerAccount) SetRatingWithJournal(rating uint32) error {
	entry, err := NewPeerJournalEntryRating(a, a.Rating)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Rating = rating

	return a.accountTracker.SaveAccount(a)
}

// GetTempRating gets the rating
func (a *PeerAccount) GetTempRating() uint32 {
	return a.TempRating
}

// SetTempRatingWithJournal sets the account's tempRating, saving the old state before changing
func (a *PeerAccount) SetTempRatingWithJournal(rating uint32) error {
	entry, err := NewPeerJournalEntryTempRating(a, a.TempRating)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.TempRating = rating

	return a.accountTracker.SaveAccount(a)
}
