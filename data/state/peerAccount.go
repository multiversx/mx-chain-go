//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. peerAccountData.proto
package state

import (
	"math/big"
)

// PeerAccount is the struct used in serialization/deserialization
type peerAccount struct {
	*baseAccount
	PeerAccountData
}

// NewEmptyPeerAccount returns an empty peerAccount
func NewEmptyPeerAccount() *peerAccount {
	return &peerAccount{
		baseAccount: &baseAccount{},
		PeerAccountData: PeerAccountData{
			Stake:           big.NewInt(0),
			AccumulatedFees: big.NewInt(0),
		},
	}
}

// NewPeerAccount creates new simple account wrapper for an PeerAccountContainer (that has just been initialized)
func NewPeerAccount(addressContainer AddressContainer) (*peerAccount, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}

	return &peerAccount{
		baseAccount: &baseAccount{
			addressContainer: addressContainer,
			dataTrieTracker:  NewTrackableDataTrie(addressContainer.Bytes(), nil),
		},
		PeerAccountData: PeerAccountData{
			Stake:           big.NewInt(0),
			AccumulatedFees: big.NewInt(0),
		},
	}, nil
}

// SetBLSPublicKey sets the account's bls public key, saving the old key before changing
func (pa *peerAccount) SetBLSPublicKey(pubKey []byte) error {
	if len(pubKey) < 1 {
		return ErrNilBLSPublicKey
	}

	pa.BLSPublicKey = pubKey
	return nil
}

// SetRewardAddress sets the account's reward address, saving the old address before changing
func (pa *peerAccount) SetRewardAddress(address []byte) error {
	if len(address) < 1 {
		return ErrEmptyAddress
	}

	pa.RewardAddress = address
	return nil
}

// SetStake sets the account's stake
func (pa *peerAccount) SetStake(stake *big.Int) error {
	if stake == nil {
		return ErrNilStake
	}

	pa.Stake = big.NewInt(0).Set(stake)
	return nil
}

// AddToAccumulatedFees sets the account's accumulated fees
func (pa *peerAccount) AddToAccumulatedFees(fees *big.Int) {
	pa.AccumulatedFees.Add(pa.AccumulatedFees, fees)
}

// IncreaseLeaderSuccessRate increases the account's number of successful signing
func (pa *peerAccount) IncreaseLeaderSuccessRate(value uint32) {
	pa.LeaderSuccessRate.NumSuccess += value
}

// DecreaseLeaderSuccessRate increases the account's number of missing signing
func (pa *peerAccount) DecreaseLeaderSuccessRate(value uint32) {
	pa.LeaderSuccessRate.NumFailure += value
}

// IncreaseValidatorSuccessRate increases the account's number of successful signing
func (pa *peerAccount) IncreaseValidatorSuccessRate(value uint32) {
	pa.ValidatorSuccessRate.NumSuccess += value
}

// DecreaseValidatorSuccessRate increases the account's number of missed signing
func (pa *peerAccount) DecreaseValidatorSuccessRate(value uint32) {
	pa.ValidatorSuccessRate.NumFailure += value
}

// IncreaseNumSelectedInSuccessBlocks sets the account's NumSelectedInSuccessBlocks
func (pa *peerAccount) IncreaseNumSelectedInSuccessBlocks() {
	pa.NumSelectedInSuccessBlocks++
}

// SetRating sets the account's rating id
func (pa *peerAccount) SetRating(rating uint32) {
	pa.Rating = rating
}

// SetTempRating sets the account's tempRating
func (pa *peerAccount) SetTempRating(rating uint32) {
	pa.TempRating = rating
}

// SetListAndIndex will update the peer's list (eligible, waiting) and the index inside it with journal
func (pa *peerAccount) SetListAndIndex(shardID uint32, list string, index uint32) {
	pa.ShardId = shardID
	pa.List = list
	pa.IndexInList = index
}

// GetList returns the list the peer is in
func (pa *peerAccount) GetList() string {
	return pa.List
}

// GetIndex returns the index in list
func (pa *peerAccount) GetIndex() uint32 {
	return pa.IndexInList
}

// IsInterfaceNil return if there is no value under the interface
func (pa *peerAccount) IsInterfaceNil() bool {
	return pa == nil
}

// ResetAtNewEpoch will reset a set of values after changing epoch
func (pa *peerAccount) ResetAtNewEpoch() {
	pa.AccumulatedFees = big.NewInt(0)
	pa.SetRating(pa.GetTempRating())
	pa.TotalLeaderSuccessRate.NumFailure += pa.LeaderSuccessRate.NumFailure
	pa.TotalLeaderSuccessRate.NumSuccess += pa.LeaderSuccessRate.NumSuccess
	pa.TotalValidatorSuccessRate.NumSuccess += pa.ValidatorSuccessRate.NumSuccess
	pa.TotalValidatorSuccessRate.NumFailure += pa.ValidatorSuccessRate.NumFailure
	pa.LeaderSuccessRate.NumFailure = 0
	pa.LeaderSuccessRate.NumSuccess = 0
	pa.ValidatorSuccessRate.NumSuccess = 0
	pa.ValidatorSuccessRate.NumFailure = 0
	pa.NumSelectedInSuccessBlocks = 0
}

// GetConsecutiveProposerMisses gets the current consecutive proposer misses
func (pa *peerAccount) GetConsecutiveProposerMisses() uint32 {
	return pa.ConsecutiveProposerMisses
}

// SetConsecutiveProposerMisses sets the account's consecutive misses as proposer
func (pa *peerAccount) SetConsecutiveProposerMisses(consecutiveMisses uint32) {
	pa.ConsecutiveProposerMisses = consecutiveMisses
}

//IncreaseNonce adds the given value to the current nonce
func (pa *peerAccount) IncreaseNonce(value uint64) {
	pa.Nonce = pa.Nonce + value
}
