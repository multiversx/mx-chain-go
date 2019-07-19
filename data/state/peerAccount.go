package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

type TimeStamp struct {
	Epoch uint64
	Round uint64
}

type TimePeriod struct {
	StartTime TimeStamp
	EndTime   TimeStamp
}

type SignRate struct {
	NrSuccess uint32
	NrFailure uint32
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

	ValidatorSuccessRate SignRate
	LeaderSuccessRate    SignRate

	Rating   uint32
	RootHash []byte
	Nonce    uint64

	addressContainer AddressContainer
	accountTracker   AccountTracker
	dataTrieTracker  DataTrieTracker
}

// NewPeerAccount creates new simple account wrapper for an PeerAccountContainer (that has just been initialized)
func NewPeerAccount(
	addressContainer AddressContainer,
	tracker AccountTracker,
	stake *big.Int,
	address []byte,
	schnorr []byte,
	bls []byte,
) (*PeerAccount, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}
	if tracker == nil {
		return nil, ErrNilAccountTracker
	}
	if stake == nil {
		return nil, ErrNilStake
	}
	if address == nil {
		return nil, ErrNilAddress
	}
	if schnorr == nil {
		return nil, ErrNilSchnorrPublicKey
	}
	if bls == nil {
		return nil, ErrNilBLSPublicKey
	}

	return &PeerAccount{
		Stake:            big.NewInt(0).Set(stake),
		Address:          address,
		SchnorrPublicKey: schnorr,
		BLSPublicKey:     bls,
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
	entry, err := NewJournalEntryNonce(a, a.Nonce)
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
	return nil
}

// SetCodeHash sets the code hash associated with the account
func (a *PeerAccount) SetCodeHash(codeHash []byte) {
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (a *PeerAccount) SetCodeHashWithJournal(codeHash []byte) error {
	return nil
}

// GetCode gets the actual code that needs to be run in the VM
func (a *PeerAccount) GetCode() []byte {
	return nil
}

// SetCode sets the actual code that needs to be run in the VM
func (a *PeerAccount) SetCode(code []byte) {
}

//------- data trie / root hash

// GetRootHash returns the root hash associated with this account
func (a *PeerAccount) GetRootHash() []byte {
	return a.RootHash
}

// SetRootHash sets the root hash associated with the account
func (a *PeerAccount) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// SetRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (a *PeerAccount) SetRootHashWithJournal(rootHash []byte) error {
	entry, err := NewBaseJournalEntryRootHash(a, a.RootHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.RootHash = rootHash

	return a.accountTracker.SaveAccount(a)
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
	entry, err := NewPeerJournalEntryAddress(a, a.Address)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Address = address

	return a.accountTracker.SaveAccount(a)
}

// SetSchnorrPublicKeyWithJournal sets the account's address, saving the old address before changing
func (a *PeerAccount) SetSchnorrPublicKeyWithJournal(pubKey []byte) error {
	entry, err := NewPeerJournalEntrySchnorrPublicKey(a, a.SchnorrPublicKey)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.SchnorrPublicKey = pubKey

	return a.accountTracker.SaveAccount(a)
}

// SetSchnorrPublicKeyWithJournal sets the account's address, saving the old address before changing
func (a *PeerAccount) SetBLSPublicKeyWithJournal(pubKey []byte) error {
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
	a.ValidatorSuccessRate.NrFailure--

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
