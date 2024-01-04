package state

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// PeerAccountHandlerMock -
type PeerAccountHandlerMock struct {
	IncreaseLeaderSuccessRateValue          uint32
	DecreaseLeaderSuccessRateValue          uint32
	IncreaseValidatorSuccessRateValue       uint32
	DecreaseValidatorSuccessRateValue       uint32
	IncreaseValidatorIgnoredSignaturesValue uint32

	IncreaseLeaderSuccessRateCalled              func(uint32)
	DecreaseLeaderSuccessRateCalled              func(uint32)
	IncreaseValidatorSuccessRateCalled           func(uint32)
	DecreaseValidatorSuccessRateCalled           func(uint32)
	IncreaseValidatorIgnoredSignaturesRateCalled func(uint32)
	SetTempRatingCalled                          func(uint32)
	GetTempRatingCalled                          func() uint32
	SetAccumulatedFeesCalled                     func(*big.Int)
	GetAccumulatedFeesCalled                     func() *big.Int
	GetConsecutiveProposerMissesCalled           func() uint32
	SetConsecutiveProposerMissesCalled           func(rating uint32)
	SetListAndIndexCalled                        func(shardID uint32, list string, index uint32)
	GetListCalled                                func() string
	GetUnStakedEpochCalled                       func() uint32
	AddressBytesCalled                           func() []byte
}

// GetUnStakedEpoch -
func (p *PeerAccountHandlerMock) GetUnStakedEpoch() uint32 {
	if p.GetUnStakedEpochCalled != nil {
		return p.GetUnStakedEpochCalled()
	}
	return common.DefaultUnstakedEpoch
}

// SetUnStakedEpoch -
func (p *PeerAccountHandlerMock) SetUnStakedEpoch(_ uint32) {
}

// GetList -
func (p *PeerAccountHandlerMock) GetList() string {
	if p.GetListCalled != nil {
		return p.GetListCalled()
	}
	return ""
}

// GetPreviousList -
func (p *PeerAccountHandlerMock) GetPreviousList() string {
	return ""
}

// GetIndexInList -
func (p *PeerAccountHandlerMock) GetIndexInList() uint32 {
	return 0
}

// GetPreviousIndexInList -
func (p *PeerAccountHandlerMock) GetPreviousIndexInList() uint32 {
	return 0
}

// GetBLSPublicKey -
func (p *PeerAccountHandlerMock) GetBLSPublicKey() []byte {
	return nil
}

// SetBLSPublicKey -
func (p *PeerAccountHandlerMock) SetBLSPublicKey([]byte) error {
	return nil
}

// GetRewardAddress -
func (p *PeerAccountHandlerMock) GetRewardAddress() []byte {
	return nil
}

// SetRewardAddress -
func (p *PeerAccountHandlerMock) SetRewardAddress([]byte) error {
	return nil
}

// GetStake -
func (p *PeerAccountHandlerMock) GetStake() *big.Int {
	return nil
}

// SetStake -
func (p *PeerAccountHandlerMock) SetStake(_ *big.Int) error {
	return nil
}

// GetAccumulatedFees -
func (p *PeerAccountHandlerMock) GetAccumulatedFees() *big.Int {
	if p.GetAccumulatedFeesCalled != nil {
		return p.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

// AddToAccumulatedFees -
func (p *PeerAccountHandlerMock) AddToAccumulatedFees(val *big.Int) {
	if p.SetAccumulatedFeesCalled != nil {
		p.SetAccumulatedFeesCalled(val)
	}
}

// GetShardId -
func (p *PeerAccountHandlerMock) GetShardId() uint32 {
	return 0
}

// IncreaseLeaderSuccessRate -
func (p *PeerAccountHandlerMock) IncreaseLeaderSuccessRate(val uint32) {
	if p.IncreaseLeaderSuccessRateCalled != nil {
		p.IncreaseLeaderSuccessRateCalled(val)
		return
	}
	p.IncreaseLeaderSuccessRateValue += val
}

// DecreaseLeaderSuccessRate -
func (p *PeerAccountHandlerMock) DecreaseLeaderSuccessRate(val uint32) {
	if p.DecreaseLeaderSuccessRateCalled != nil {
		p.DecreaseLeaderSuccessRateCalled(val)
		return
	}
	p.DecreaseLeaderSuccessRateValue += val
}

// IncreaseValidatorSuccessRate -
func (p *PeerAccountHandlerMock) IncreaseValidatorSuccessRate(val uint32) {
	if p.IncreaseValidatorSuccessRateCalled != nil {
		p.IncreaseValidatorSuccessRateCalled(val)
		return
	}
	p.IncreaseValidatorSuccessRateValue += val
}

// DecreaseValidatorSuccessRate -
func (p *PeerAccountHandlerMock) DecreaseValidatorSuccessRate(val uint32) {
	if p.DecreaseValidatorSuccessRateCalled != nil {
		p.DecreaseValidatorSuccessRateCalled(val)
		return
	}
	p.DecreaseValidatorSuccessRateValue += val
}

// IncreaseValidatorIgnoredSignaturesRate -
func (p *PeerAccountHandlerMock) IncreaseValidatorIgnoredSignaturesRate(val uint32) {
	if p.IncreaseValidatorIgnoredSignaturesRateCalled != nil {
		p.IncreaseValidatorIgnoredSignaturesRateCalled(val)
		return
	}
	p.IncreaseValidatorIgnoredSignaturesValue += val
}

// GetNumSelectedInSuccessBlocks -
func (p *PeerAccountHandlerMock) GetNumSelectedInSuccessBlocks() uint32 {
	return 0
}

// IncreaseNumSelectedInSuccessBlocks -
func (p *PeerAccountHandlerMock) IncreaseNumSelectedInSuccessBlocks() {

}

// GetLeaderSuccessRate -
func (p *PeerAccountHandlerMock) GetLeaderSuccessRate() state.SignRate {
	return &SignRate{}
}

// GetValidatorSuccessRate -
func (p *PeerAccountHandlerMock) GetValidatorSuccessRate() state.SignRate {
	return &SignRate{}
}

// GetValidatorIgnoredSignaturesRate -
func (p *PeerAccountHandlerMock) GetValidatorIgnoredSignaturesRate() uint32 {
	return 0
}

// GetTotalLeaderSuccessRate -
func (p *PeerAccountHandlerMock) GetTotalLeaderSuccessRate() state.SignRate {
	return &SignRate{}
}

// GetTotalValidatorSuccessRate -
func (p *PeerAccountHandlerMock) GetTotalValidatorSuccessRate() state.SignRate {
	return &SignRate{}
}

// GetTotalValidatorIgnoredSignaturesRate -
func (p *PeerAccountHandlerMock) GetTotalValidatorIgnoredSignaturesRate() uint32 {
	return 0
}

// GetRating -
func (p *PeerAccountHandlerMock) GetRating() uint32 {
	return 0
}

// SetRating -
func (p *PeerAccountHandlerMock) SetRating(uint32) {

}

// GetTempRating -
func (p *PeerAccountHandlerMock) GetTempRating() uint32 {
	if p.GetTempRatingCalled != nil {
		return p.GetTempRatingCalled()
	}
	return 0
}

// SetTempRating -
func (p *PeerAccountHandlerMock) SetTempRating(val uint32) {
	if p.SetTempRatingCalled != nil {
		p.SetTempRatingCalled(val)
	}
}

// ResetAtNewEpoch -
func (p *PeerAccountHandlerMock) ResetAtNewEpoch() {
}

// AddressBytes -
func (p *PeerAccountHandlerMock) AddressBytes() []byte {
	if p.AddressBytesCalled != nil {
		return p.AddressBytesCalled()
	}

	return nil
}

// IncreaseNonce -
func (p *PeerAccountHandlerMock) IncreaseNonce(_ uint64) {
}

// GetNonce -
func (p *PeerAccountHandlerMock) GetNonce() uint64 {
	return 0
}

// SetCode -
func (p *PeerAccountHandlerMock) SetCode(_ []byte) {

}

// GetCode -
func (p *PeerAccountHandlerMock) GetCode() []byte {
	return nil
}

// SetCodeHash -
func (p *PeerAccountHandlerMock) SetCodeHash(_ []byte) {

}

// GetCodeHash -
func (p *PeerAccountHandlerMock) GetCodeHash() []byte {
	return nil
}

// SetRootHash -
func (p *PeerAccountHandlerMock) SetRootHash([]byte) {

}

// GetRootHash -
func (p *PeerAccountHandlerMock) GetRootHash() []byte {
	return nil
}

// SetDataTrie -
func (p *PeerAccountHandlerMock) SetDataTrie(_ common.Trie) {

}

// DataTrie -
func (p *PeerAccountHandlerMock) DataTrie() common.Trie {
	return nil
}

// GetConsecutiveProposerMisses -
func (p *PeerAccountHandlerMock) GetConsecutiveProposerMisses() uint32 {
	if p.GetConsecutiveProposerMissesCalled != nil {
		return p.GetConsecutiveProposerMissesCalled()
	}
	return 0
}

// SetConsecutiveProposerMisses -
func (p *PeerAccountHandlerMock) SetConsecutiveProposerMisses(consecutiveMisses uint32) {
	if p.SetConsecutiveProposerMissesCalled != nil {
		p.SetConsecutiveProposerMissesCalled(consecutiveMisses)
	}
}

// SetListAndIndex -
func (p *PeerAccountHandlerMock) SetListAndIndex(shardID uint32, list string, index uint32, _ bool) {
	if p.SetListAndIndexCalled != nil {
		p.SetListAndIndexCalled(shardID, list, index)
	}
}

// IsInterfaceNil -
func (p *PeerAccountHandlerMock) IsInterfaceNil() bool {
	return false
}
