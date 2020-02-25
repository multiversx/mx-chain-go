package accounts

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// PeerAccount is the struct used in serialization/deserialization
type peerAccount struct {
	*baseAccount
	BLSPublicKey     []byte
	SchnorrPublicKey []byte
	RewardAddress    []byte
	Stake            *big.Int
	AccumulatedFees  *big.Int

	JailTime      state.TimePeriod
	PastJailTimes []state.TimePeriod

	CurrentShardId    uint32
	NextShardId       uint32
	NodeInWaitingList bool
	UnStakedNonce     uint64

	ValidatorSuccessRate       state.SignRate
	LeaderSuccessRate          state.SignRate
	NumSelectedInSuccessBlocks uint32

	Rating     uint32
	TempRating uint32
}

// NewPeerAccount creates new simple account wrapper for an PeerAccountContainer (that has just been initialized)
func NewPeerAccount(addressContainer state.AddressContainer) (*peerAccount, error) {
	if addressContainer == nil {
		return nil, state.ErrNilAddressContainer
	}

	return &peerAccount{
		baseAccount: &baseAccount{
			addressContainer: addressContainer,
			dataTrieTracker:  state.NewTrackableDataTrie(addressContainer.Bytes(), nil),
		},
		Stake: big.NewInt(0),
	}, nil
}

func (pa *peerAccount) GetBLSPublicKey() []byte {
	return pa.BLSPublicKey
}

// SetBLSPublicKey sets the account's bls public key, saving the old key before changing
func (pa *peerAccount) SetBLSPublicKey(pubKey []byte) {
	pa.BLSPublicKey = pubKey
}

func (pa *peerAccount) GetSchnorrPublicKey() []byte {
	return pa.SchnorrPublicKey
}

// SetSchnorrPublicKey sets the account's public key, saving the old key before changing
func (pa *peerAccount) SetSchnorrPublicKey(pubKey []byte) {
	pa.SchnorrPublicKey = pubKey
}

func (pa *peerAccount) GetRewardAddress() []byte {
	return pa.RewardAddress
}

// SetRewardAddress sets the account's reward address, saving the old address before changing
func (pa *peerAccount) SetRewardAddress(address []byte) {
	pa.RewardAddress = address
}

func (pa *peerAccount) GetStake() *big.Int {
	return pa.Stake
}

// SetStake sets the account's stake, saving the old stake before changing
func (pa *peerAccount) SetStake(stake *big.Int) {
	pa.Stake = stake
}

func (pa *peerAccount) GetAccumulatedFees() *big.Int {
	return pa.AccumulatedFees
}

func (pa *peerAccount) SetAccumulatedFees(fees *big.Int) {
	pa.AccumulatedFees = fees
}

func (pa *peerAccount) GetJailTime() state.TimePeriod {
	return pa.JailTime
}

// SetJailTime sets the account's jail time, saving the old state before changing
func (pa *peerAccount) SetJailTime(jailTime state.TimePeriod) {
	pa.JailTime = jailTime
}

func (pa *peerAccount) GetCurrentShardId() uint32 {
	return pa.CurrentShardId
}

// SetCurrentShardId sets the account's shard id, saving the old state before changing
func (pa *peerAccount) SetCurrentShardId(shId uint32) {
	pa.CurrentShardId = shId
}

func (pa *peerAccount) GetNextShardId() uint32 {
	return pa.NextShardId
}

// SetNextShardId sets the account's shard id, saving the old state before changing
func (pa *peerAccount) SetNextShardId(shId uint32) {
	pa.NextShardId = shId
}

func (pa *peerAccount) GetNodeInWaitingList() bool {
	return pa.NodeInWaitingList
}

// SetNodeInWaitingList sets the account's nodes status whether in waiting list, saving the old state before
func (pa *peerAccount) SetNodeInWaitingList(nodeInWaitingList bool) {
	pa.NodeInWaitingList = nodeInWaitingList
}

func (pa *peerAccount) GetUnStakedNonce() uint64 {
	return pa.UnStakedNonce
}

// SetUnStakedNonce sets the account's shard id, saving the old state before changing
func (pa *peerAccount) SetUnStakedNonce(nonce uint64) {
	pa.UnStakedNonce = nonce
}

// IncreaseLeaderSuccessRate increases the account's number of successful signing,
// saving the old state before changing
func (pa *peerAccount) IncreaseLeaderSuccessRate(value uint32) {
	pa.LeaderSuccessRate.NrSuccess += value
}

// DecreaseLeaderSuccessRate increases the account's number of missing signing,
// saving the old state before changing
func (pa *peerAccount) DecreaseLeaderSuccessRate(value uint32) {
	pa.LeaderSuccessRate.NrFailure += value
}

// IncreaseValidatorSuccessRate increases the account's number of successful signing,
// saving the old state before changing
func (pa *peerAccount) IncreaseValidatorSuccessRate(value uint32) {
	pa.ValidatorSuccessRate.NrSuccess += value
}

// DecreaseValidatorSuccessRate increases the account's number of missed signing,
// saving the old state before changing
func (pa *peerAccount) DecreaseValidatorSuccessRate(value uint32) {
	pa.ValidatorSuccessRate.NrFailure += value
}

func (pa *peerAccount) GetNumSelectedInSuccessBlocks() uint32 {
	return pa.NumSelectedInSuccessBlocks
}

func (pa *peerAccount) SetNumSelectedInSuccessBlocks(num uint32) {
	pa.NumSelectedInSuccessBlocks = num
}

// GetRating gets the rating
func (pa *peerAccount) GetRating() uint32 {
	return pa.Rating
}

// SetRating sets the account's rating id, saving the old state before changing
func (pa *peerAccount) SetRating(rating uint32) {
	pa.Rating = rating
}

// GetTempRating gets the rating
func (pa *peerAccount) GetTempRating() uint32 {
	return pa.TempRating
}

// SetTempRating sets the account's tempRating, saving the old state before changing
func (pa *peerAccount) SetTempRating(rating uint32) {
	pa.TempRating = rating
}

// IsInterfaceNil return if there is no value under the interface
func (pa *peerAccount) IsInterfaceNil() bool {
	return pa == nil
}

//// ResetAtNewEpoch will reset a set of values after changing epoch
//func (pa *PeerAccount) ResetAtNewEpoch() error {
//	entryAccFee, err := NewPeerJournalEntryAccumulatedFees(pa, pa.AccumulatedFees)
//	if err != nil {
//		return err
//	}
//
//	pa.accountTracker.Journalize(entryAccFee)
//	pa.AccumulatedFees = big.NewInt(0)
//
//	err = pa.accountTracker.SaveAccount(pa)
//	if err != nil {
//		return err
//	}
//
//	err = pa.SetRatingWithJournal(pa.GetTempRating())
//	if err != nil {
//		return err
//	}
//
//	entryLeaderRate, err := NewPeerJournalEntryLeaderSuccessRate(pa, pa.LeaderSuccessRate)
//	if err != nil {
//		return err
//	}
//
//	pa.accountTracker.Journalize(entryLeaderRate)
//	pa.LeaderSuccessRate.NrFailure = 0
//	pa.LeaderSuccessRate.NrSuccess = 0
//
//	err = pa.accountTracker.SaveAccount(pa)
//	if err != nil {
//		return err
//	}
//
//	entryValidatorRate, err := NewPeerJournalEntryValidatorSuccessRate(pa, pa.ValidatorSuccessRate)
//	if err != nil {
//		return err
//	}
//
//	pa.accountTracker.Journalize(entryValidatorRate)
//	pa.ValidatorSuccessRate.NrSuccess = 0
//	pa.ValidatorSuccessRate.NrFailure = 0
//
//	err = pa.accountTracker.SaveAccount(pa)
//	if err != nil {
//		return err
//	}
//
//	entry, err := NewPeerJournalEntryNumSelectedInSuccessBlocks(pa, pa.NumSelectedInSuccessBlocks)
//	if err != nil {
//		return err
//	}
//
//	pa.accountTracker.Journalize(entry)
//	pa.NumSelectedInSuccessBlocks = 0
//
//	return pa.accountTracker.SaveAccount(pa)
//}
//
