package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type PeerAccountHandlerMock struct {
	IncreaseLeaderSuccessRateCalled    func(uint32)
	DecreaseLeaderSuccessRateCalled    func(uint32)
	IncreaseValidatorSuccessRateCalled func(uint32)
	DecreaseValidatorSuccessRateCalled func(uint32)
	SetTempRatingCalled                func(uint32)
	GetTempRatingCalled                func() uint32
	SetAccumulatedFeesCalled           func(*big.Int)
	GetAccumulatedFeesCalled           func() *big.Int
}

func (p *PeerAccountHandlerMock) GetBLSPublicKey() []byte {
	return nil
}

func (p *PeerAccountHandlerMock) SetBLSPublicKey([]byte) {

}

func (p *PeerAccountHandlerMock) GetSchnorrPublicKey() []byte {
	return nil
}

func (p *PeerAccountHandlerMock) SetSchnorrPublicKey([]byte) {

}

func (p *PeerAccountHandlerMock) GetRewardAddress() []byte {
	return nil
}

func (p *PeerAccountHandlerMock) SetRewardAddress([]byte) error {
	return nil
}

func (p *PeerAccountHandlerMock) GetStake() *big.Int {
	return nil
}

func (p *PeerAccountHandlerMock) SetStake(*big.Int) {

}

func (p *PeerAccountHandlerMock) GetAccumulatedFees() *big.Int {
	if p.GetAccumulatedFeesCalled != nil {
		p.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

func (p *PeerAccountHandlerMock) SetAccumulatedFees(val *big.Int) {
	if p.SetAccumulatedFeesCalled != nil {
		p.SetAccumulatedFeesCalled(val)
	}
}

func (p *PeerAccountHandlerMock) GetJailTime() state.TimePeriod {
	return state.TimePeriod{}
}

func (p *PeerAccountHandlerMock) SetJailTime(state.TimePeriod) {

}

func (p *PeerAccountHandlerMock) GetCurrentShardId() uint32 {
	return 0
}

func (p *PeerAccountHandlerMock) SetCurrentShardId(uint32) {

}

func (p *PeerAccountHandlerMock) GetNextShardId() uint32 {
	return 0
}

func (p *PeerAccountHandlerMock) SetNextShardId(uint32) {

}

func (p *PeerAccountHandlerMock) GetNodeInWaitingList() bool {
	return false
}

func (p *PeerAccountHandlerMock) SetNodeInWaitingList(bool) {

}

func (p *PeerAccountHandlerMock) GetUnStakedNonce() uint64 {
	return 0
}

func (p *PeerAccountHandlerMock) SetUnStakedNonce(uint64) {

}

func (p *PeerAccountHandlerMock) IncreaseLeaderSuccessRate(val uint32) {
	if p.IncreaseLeaderSuccessRateCalled != nil {
		p.IncreaseLeaderSuccessRateCalled(val)
	}
}

func (p *PeerAccountHandlerMock) DecreaseLeaderSuccessRate(val uint32) {
	if p.DecreaseLeaderSuccessRateCalled != nil {
		p.DecreaseLeaderSuccessRateCalled(val)
	}
}

func (p *PeerAccountHandlerMock) IncreaseValidatorSuccessRate(val uint32) {
	if p.IncreaseValidatorSuccessRateCalled != nil {
		p.IncreaseValidatorSuccessRateCalled(val)
	}
}

func (p *PeerAccountHandlerMock) DecreaseValidatorSuccessRate(val uint32) {
	if p.DecreaseValidatorSuccessRateCalled != nil {
		p.DecreaseValidatorSuccessRateCalled(val)
	}
}

func (p *PeerAccountHandlerMock) GetNumSelectedInSuccessBlocks() uint32 {
	return 0
}

func (p *PeerAccountHandlerMock) SetNumSelectedInSuccessBlocks(uint32) {

}

func (p *PeerAccountHandlerMock) GetLeaderSuccessRate() state.SignRate {
	return state.SignRate{}
}

func (p *PeerAccountHandlerMock) GetValidatorSuccessRate() state.SignRate {
	return state.SignRate{}
}

func (p *PeerAccountHandlerMock) GetRating() uint32 {
	return 0
}

func (p *PeerAccountHandlerMock) SetRating(uint32) {

}

func (p *PeerAccountHandlerMock) GetTempRating() uint32 {
	if p.GetTempRatingCalled != nil {
		return p.GetTempRatingCalled()
	}
	return 0
}

func (p *PeerAccountHandlerMock) SetTempRating(val uint32) {
	if p.SetTempRatingCalled != nil {
		p.SetTempRatingCalled(val)
	}
}

func (p *PeerAccountHandlerMock) ResetAtNewEpoch() error {
	return nil
}

func (p *PeerAccountHandlerMock) AddressContainer() state.AddressContainer {
	return nil
}

func (p *PeerAccountHandlerMock) SetNonce(nonce uint64) {

}

func (p *PeerAccountHandlerMock) GetNonce() uint64 {
	return 0
}

func (p *PeerAccountHandlerMock) SetCode(code []byte) {

}

func (p *PeerAccountHandlerMock) GetCode() []byte {
	return nil
}

func (p *PeerAccountHandlerMock) SetCodeHash([]byte) {

}

func (p *PeerAccountHandlerMock) GetCodeHash() []byte {
	return nil
}

func (p *PeerAccountHandlerMock) SetRootHash([]byte) {

}

func (p *PeerAccountHandlerMock) GetRootHash() []byte {
	return nil
}

func (p *PeerAccountHandlerMock) SetDataTrie(trie data.Trie) {

}

func (p *PeerAccountHandlerMock) DataTrie() data.Trie {
	return nil
}

func (p *PeerAccountHandlerMock) DataTrieTracker() state.DataTrieTracker {
	return nil
}

func (p *PeerAccountHandlerMock) IsInterfaceNil() bool {
	return false
}
