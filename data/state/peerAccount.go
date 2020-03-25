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

// SetSchnorrPublicKey sets the account's public key, saving the old key before changing
func (pa *peerAccount) SetSchnorrPublicKey(pubKey []byte) error {
	if len(pubKey) < 1 {
		return ErrNilSchnorrPublicKey
	}

	pa.SchnorrPublicKey = pubKey
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

// SetAccumulatedFees sets the account's accumulated fees
func (pa *peerAccount) SetAccumulatedFees(fees *big.Int) {
	pa.AccumulatedFees = big.NewInt(0).Set(fees)
}

// SetJailTime sets the account's jail time
func (pa *peerAccount) SetJailTime(jailTime TimePeriod) {
	pa.JailTime = jailTime
}

// SetCurrentShardId sets the account's shard id
func (pa *peerAccount) SetCurrentShardId(shId uint32) {
	pa.CurrentShardId = shId
}

// SetNextShardId sets the account's shard id
func (pa *peerAccount) SetNextShardId(shId uint32) {
	pa.NextShardId = shId
}

// SetNodeInWaitingList sets the account's nodes status whether in waiting list
func (pa *peerAccount) SetNodeInWaitingList(nodeInWaitingList bool) {
	pa.NodeInWaitingList = nodeInWaitingList
}

// SetUnStakedNonce sets the account's shard id
func (pa *peerAccount) SetUnStakedNonce(nonce uint64) {
	pa.UnStakedNonce = nonce
}

// IncreaseLeaderSuccessRate increases the account's number of successful signing
func (pa *peerAccount) IncreaseLeaderSuccessRate(value uint32) {
	pa.LeaderSuccessRate.NrSuccess += value
}

// DecreaseLeaderSuccessRate increases the account's number of missing signing
func (pa *peerAccount) DecreaseLeaderSuccessRate(value uint32) {
	pa.LeaderSuccessRate.NrFailure += value
}

// IncreaseValidatorSuccessRate increases the account's number of successful signing
func (pa *peerAccount) IncreaseValidatorSuccessRate(value uint32) {
	pa.ValidatorSuccessRate.NrSuccess += value
}

// DecreaseValidatorSuccessRate increases the account's number of missed signing
func (pa *peerAccount) DecreaseValidatorSuccessRate(value uint32) {
	pa.ValidatorSuccessRate.NrFailure += value
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

// IsInterfaceNil return if there is no value under the interface
func (pa *peerAccount) IsInterfaceNil() bool {
	return pa == nil
}

// ResetAtNewEpoch will reset a set of values after changing epoch
func (pa *peerAccount) ResetAtNewEpoch() error {
	pa.AccumulatedFees = big.NewInt(0)
	pa.SetRating(pa.GetTempRating())
	pa.LeaderSuccessRate.NrFailure = 0
	pa.LeaderSuccessRate.NrSuccess = 0
	pa.ValidatorSuccessRate.NrSuccess = 0
	pa.ValidatorSuccessRate.NrFailure = 0
	pa.NumSelectedInSuccessBlocks = 0

	return nil
}

//IncreaseNonce adds the given value to the current nonce
func (pa *peerAccount) IncreaseNonce(val uint64) {
	pa.Nonce = pa.Nonce + val
}
