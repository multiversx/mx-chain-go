//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. peerAccountData.proto
package state

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
)

// PeerAccount is the struct used in serialization/deserialization
type peerAccount struct {
	PeerAccountData
}

// TODO: replace NewEmptyPeerAccount with GetPeerAccountFromBytes

// NewEmptyPeerAccount returns an empty peerAccount
func NewEmptyPeerAccount() *peerAccount {
	return &peerAccount{
		PeerAccountData: PeerAccountData{
			AccumulatedFees: big.NewInt(0),
			UnStakedEpoch:   common.DefaultUnstakedEpoch,
		},
	}
}

// NewPeerAccount creates a new instance of peerAccount
func NewPeerAccount(address []byte) (*peerAccount, error) {
	if len(address) == 0 {
		return nil, ErrNilAddress
	}

	return &peerAccount{
		PeerAccountData: PeerAccountData{
			BLSPublicKey:    address,
			AccumulatedFees: big.NewInt(0),
			UnStakedEpoch:   common.DefaultUnstakedEpoch,
		},
	}, nil
}

// AddressBytes returns the account's address
func (pa *peerAccount) AddressBytes() []byte {
	return pa.BLSPublicKey
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

// IncreaseValidatorIgnoredSignaturesRate increases the account's number of ignored signatures in successful blocks
func (pa *peerAccount) IncreaseValidatorIgnoredSignaturesRate(value uint32) {
	pa.ValidatorIgnoredSignaturesRate += value
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

// SetUnStakedEpoch updates the unstaked epoch for the validator
func (pa *peerAccount) SetUnStakedEpoch(epoch uint32) {
	pa.UnStakedEpoch = epoch
}

// ResetAtNewEpoch will reset a set of values after changing epoch
func (pa *peerAccount) ResetAtNewEpoch() {
	pa.AccumulatedFees = big.NewInt(0)
	pa.SetRating(pa.GetTempRating())
	pa.TotalLeaderSuccessRate.NumFailure += pa.LeaderSuccessRate.NumFailure
	pa.TotalLeaderSuccessRate.NumSuccess += pa.LeaderSuccessRate.NumSuccess
	pa.TotalValidatorSuccessRate.NumSuccess += pa.ValidatorSuccessRate.NumSuccess
	pa.TotalValidatorSuccessRate.NumFailure += pa.ValidatorSuccessRate.NumFailure
	pa.TotalValidatorIgnoredSignaturesRate += pa.ValidatorIgnoredSignaturesRate
	pa.LeaderSuccessRate.NumFailure = 0
	pa.LeaderSuccessRate.NumSuccess = 0
	pa.ValidatorSuccessRate.NumSuccess = 0
	pa.ValidatorSuccessRate.NumFailure = 0
	pa.ValidatorIgnoredSignaturesRate = 0
	pa.NumSelectedInSuccessBlocks = 0
}

// SetConsecutiveProposerMisses sets the account's consecutive misses as proposer
func (pa *peerAccount) SetConsecutiveProposerMisses(consecutiveMisses uint32) {
	pa.ConsecutiveProposerMisses = consecutiveMisses
}

// IncreaseNonce adds the given value to the current nonce
func (pa *peerAccount) IncreaseNonce(value uint64) {
	pa.Nonce = pa.Nonce + value
}

// IsInterfaceNil return if there is no value under the interface
func (pa *peerAccount) IsInterfaceNil() bool {
	return pa == nil
}
