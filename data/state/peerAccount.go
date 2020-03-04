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
	RewardAddress    []byte
	Stake            *big.Int
	AccumulatedFees  *big.Int

	JailTime      TimePeriod
	PastJailTimes []TimePeriod

	CurrentShardId    uint32
	NextShardId       uint32
	NodeInWaitingList bool
	UnStakedNonce     uint64

	IndexInList int
	List        string

	ValidatorSuccessRate       SignRate
	LeaderSuccessRate          SignRate
	NumSelectedInSuccessBlocks uint32

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
		AccumulatedFees:  big.NewInt(0),
		Stake:            big.NewInt(0),
		addressContainer: addressContainer,
		accountTracker:   tracker,
		dataTrieTracker:  NewTrackableDataTrie(addressContainer.Bytes(), nil),
	}, nil
}

// IsInterfaceNil return if there is no value under the interface
func (pa *PeerAccount) IsInterfaceNil() bool {
	return pa == nil
}

// AddressContainer returns the address associated with the account
func (pa *PeerAccount) AddressContainer() AddressContainer {
	return pa.addressContainer
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (pa *PeerAccount) SetNonceWithJournal(nonce uint64) error {
	entry, err := NewBaseJournalEntryNonce(pa, pa.Nonce)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.Nonce = nonce

	return pa.accountTracker.SaveAccount(pa)
}

//SetNonce saves the nonce to the account
func (pa *PeerAccount) SetNonce(nonce uint64) {
	pa.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (pa *PeerAccount) GetNonce() uint64 {
	return pa.Nonce
}

// GetCodeHash returns the code hash associated with this account
func (pa *PeerAccount) GetCodeHash() []byte {
	return pa.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (pa *PeerAccount) SetCodeHash(codeHash []byte) {
	pa.CodeHash = codeHash
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (pa *PeerAccount) SetCodeHashWithJournal(codeHash []byte) error {
	entry, err := NewBaseJournalEntryCodeHash(pa, pa.CodeHash)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.CodeHash = codeHash

	return pa.accountTracker.SaveAccount(pa)
}

// GetCode gets the actual code that needs to be run in the VM
func (pa *PeerAccount) GetCode() []byte {
	return pa.code
}

// SetCode sets the actual code that needs to be run in the VM
func (pa *PeerAccount) SetCode(code []byte) {
	pa.code = code
}

// GetRootHash returns the root hash associated with this account
func (pa *PeerAccount) GetRootHash() []byte {
	return pa.RootHash
}

// SetRootHash sets the root hash associated with the account
func (pa *PeerAccount) SetRootHash(roothash []byte) {
	pa.RootHash = roothash
}

// DataTrie returns the trie that holds the current account's data
func (pa *PeerAccount) DataTrie() data.Trie {
	return pa.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (pa *PeerAccount) SetDataTrie(trie data.Trie) {
	pa.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (pa *PeerAccount) DataTrieTracker() DataTrieTracker {
	return pa.dataTrieTracker
}

// SetRewardAddressWithJournal sets the account's reward address, saving the old address before changing
func (pa *PeerAccount) SetRewardAddressWithJournal(address []byte) error {
	if len(address) < 1 {
		return ErrEmptyAddress
	}

	entry, err := NewPeerJournalEntryAddress(pa, pa.RewardAddress)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.RewardAddress = address

	return pa.accountTracker.SaveAccount(pa)
}

// SetSchnorrPublicKeyWithJournal sets the account's public key, saving the old key before changing
func (pa *PeerAccount) SetSchnorrPublicKeyWithJournal(pubKey []byte) error {
	if len(pubKey) < 1 {
		return ErrNilSchnorrPublicKey
	}

	entry, err := NewPeerJournalEntrySchnorrPublicKey(pa, pa.SchnorrPublicKey)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.SchnorrPublicKey = pubKey

	return pa.accountTracker.SaveAccount(pa)
}

// SetBLSPublicKeyWithJournal sets the account's bls public key, saving the old key before changing
func (pa *PeerAccount) SetBLSPublicKeyWithJournal(pubKey []byte) error {
	if len(pubKey) < 1 {
		return ErrNilBLSPublicKey
	}

	entry, err := NewPeerJournalEntryBLSPublicKey(pa, pa.BLSPublicKey)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.BLSPublicKey = pubKey

	return pa.accountTracker.SaveAccount(pa)
}

// SetStakeWithJournal sets the account's stake, saving the old stake before changing
func (pa *PeerAccount) SetStakeWithJournal(stake *big.Int) error {
	if stake == nil {
		return ErrNilStake
	}

	entry, err := NewPeerJournalEntryStake(pa, pa.Stake)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.Stake = stake

	return pa.accountTracker.SaveAccount(pa)
}

// SetJailTimeWithJournal sets the account's jail time, saving the old state before changing
func (pa *PeerAccount) SetJailTimeWithJournal(jailTime TimePeriod) error {
	entry, err := NewPeerJournalEntryJailTime(pa, pa.JailTime)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.JailTime = jailTime

	return pa.accountTracker.SaveAccount(pa)
}

// SetUnStakedNonceWithJournal sets the account's shard id, saving the old state before changing
func (pa *PeerAccount) SetUnStakedNonceWithJournal(nonce uint64) error {
	entry, err := NewPeerJournalEntryUnStakedNonce(pa, pa.UnStakedNonce)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.UnStakedNonce = nonce

	return pa.accountTracker.SaveAccount(pa)
}

// SetCurrentShardIdWithJournal sets the account's shard id, saving the old state before changing
func (pa *PeerAccount) SetCurrentShardIdWithJournal(shId uint32) error {
	entry, err := NewPeerJournalEntryCurrentShardId(pa, pa.CurrentShardId)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.CurrentShardId = shId

	return pa.accountTracker.SaveAccount(pa)
}

// SetNextShardIdWithJournal sets the account's shard id, saving the old state before changing
func (pa *PeerAccount) SetNextShardIdWithJournal(shId uint32) error {
	entry, err := NewPeerJournalEntryNextShardId(pa, pa.NextShardId)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.NextShardId = shId

	return pa.accountTracker.SaveAccount(pa)
}

// SetNodeInWaitingListWithJournal sets the account's nodes status whether in waiting list, saving the old state before
func (pa *PeerAccount) SetNodeInWaitingListWithJournal(nodeInWaitingList bool) error {
	entry, err := NewPeerJournalEntryInWaitingList(pa, pa.NodeInWaitingList)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.NodeInWaitingList = nodeInWaitingList

	return pa.accountTracker.SaveAccount(pa)
}

// IncreaseValidatorSuccessRateWithJournal increases the account's number of successful signing,
// saving the old state before changing
func (pa *PeerAccount) IncreaseValidatorSuccessRateWithJournal(value uint32) error {
	entry, err := NewPeerJournalEntryValidatorSuccessRate(pa, pa.ValidatorSuccessRate)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.ValidatorSuccessRate.NrSuccess += value

	return pa.accountTracker.SaveAccount(pa)
}

// IncreaseNumSelectedInSuccessBlocks increases the counter for number of selection in successful blocks
func (pa *PeerAccount) IncreaseNumSelectedInSuccessBlocks() error {
	entry, err := NewPeerJournalEntryNumSelectedInSuccessBlocks(pa, pa.NumSelectedInSuccessBlocks)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.NumSelectedInSuccessBlocks += 1

	return pa.accountTracker.SaveAccount(pa)
}

// DecreaseValidatorSuccessRateWithJournal increases the account's number of missed signing,
// saving the old state before changing
func (pa *PeerAccount) DecreaseValidatorSuccessRateWithJournal(value uint32) error {
	entry, err := NewPeerJournalEntryValidatorSuccessRate(pa, pa.ValidatorSuccessRate)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.ValidatorSuccessRate.NrFailure += value

	return pa.accountTracker.SaveAccount(pa)
}

// IncreaseLeaderAccumulatedFees increases the account's accumulated fees
func (pa *PeerAccount) AddToAccumulatedFees(value *big.Int) error {
	if value.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	entry, err := NewPeerJournalEntryAccumulatedFees(pa, pa.AccumulatedFees)
	if err != nil {
		return err
	}

	newAccumulatedFees := big.NewInt(0).Add(pa.AccumulatedFees, value)
	pa.accountTracker.Journalize(entry)
	pa.AccumulatedFees = newAccumulatedFees

	return pa.accountTracker.SaveAccount(pa)
}

// ResetAtNewEpoch will reset a set of values after changing epoch
func (pa *PeerAccount) ResetAtNewEpoch() error {
	entryAccFee, err := NewPeerJournalEntryAccumulatedFees(pa, pa.AccumulatedFees)
	if err != nil {
		return err
	}

	newAccumulatedFees := big.NewInt(0)
	pa.accountTracker.Journalize(entryAccFee)
	pa.AccumulatedFees = newAccumulatedFees

	err = pa.accountTracker.SaveAccount(pa)
	if err != nil {
		return err
	}

	err = pa.SetRatingWithJournal(pa.GetTempRating())
	if err != nil {
		return err
	}

	entryLeaderRate, err := NewPeerJournalEntryLeaderSuccessRate(pa, pa.LeaderSuccessRate)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entryLeaderRate)
	pa.LeaderSuccessRate.NrFailure = 0
	pa.LeaderSuccessRate.NrSuccess = 0

	err = pa.accountTracker.SaveAccount(pa)
	if err != nil {
		return err
	}

	entryValidatorRate, err := NewPeerJournalEntryValidatorSuccessRate(pa, pa.ValidatorSuccessRate)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entryValidatorRate)
	pa.ValidatorSuccessRate.NrSuccess = 0
	pa.ValidatorSuccessRate.NrFailure = 0

	err = pa.accountTracker.SaveAccount(pa)
	if err != nil {
		return err
	}

	entry, err := NewPeerJournalEntryNumSelectedInSuccessBlocks(pa, pa.NumSelectedInSuccessBlocks)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.NumSelectedInSuccessBlocks = 0

	return pa.accountTracker.SaveAccount(pa)
}

// IncreaseLeaderSuccessRateWithJournal increases the account's number of successful signing,
// saving the old state before changing
func (pa *PeerAccount) IncreaseLeaderSuccessRateWithJournal(value uint32) error {
	entry, err := NewPeerJournalEntryLeaderSuccessRate(pa, pa.LeaderSuccessRate)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.LeaderSuccessRate.NrSuccess += value

	return pa.accountTracker.SaveAccount(pa)
}

// DecreaseLeaderSuccessRateWithJournal increases the account's number of missing signing,
// saving the old state before changing
func (pa *PeerAccount) DecreaseLeaderSuccessRateWithJournal(value uint32) error {
	entry, err := NewPeerJournalEntryLeaderSuccessRate(pa, pa.LeaderSuccessRate)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.LeaderSuccessRate.NrFailure += value

	return pa.accountTracker.SaveAccount(pa)
}

// GetRating gets the rating
func (pa *PeerAccount) GetRating() uint32 {
	return pa.Rating
}

// SetListAndIndexWithJournal will update the peer's list (eligible, waiting) and the index inside it with journal
func (pa *PeerAccount) SetListAndIndexWithJournal(shardID uint32, list string, index int) error {
	entry, err := NewPeerJournalEntryListIndex(pa, pa.CurrentShardId, pa.List, pa.IndexInList)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.CurrentShardId = shardID
	pa.List = list
	pa.IndexInList = index

	return pa.accountTracker.SaveAccount(pa)
}

// SetRatingWithJournal sets the account's rating id, saving the old state before changing
func (pa *PeerAccount) SetRatingWithJournal(rating uint32) error {
	entry, err := NewPeerJournalEntryRating(pa, pa.Rating)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.Rating = rating

	return pa.accountTracker.SaveAccount(pa)
}

// GetTempRating gets the rating
func (pa *PeerAccount) GetTempRating() uint32 {
	return pa.TempRating
}

// SetTempRatingWithJournal sets the account's tempRating, saving the old state before changing
func (pa *PeerAccount) SetTempRatingWithJournal(rating uint32) error {
	entry, err := NewPeerJournalEntryTempRating(pa, pa.TempRating)
	if err != nil {
		return err
	}

	pa.accountTracker.Journalize(entry)
	pa.TempRating = rating

	return pa.accountTracker.SaveAccount(pa)
}
