package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/track"
)

// BlockTrackerHandlerMock -
type BlockTrackerHandlerMock struct {
	GetSelfHeadersCalled          func(headerHandler data.HeaderHandler) []*track.HeaderInfo
	ComputeCrossInfoCalled        func(headers []data.HeaderHandler)
	ComputeLongestSelfChainCalled func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte)
	SortHeadersFromNonceCalled    func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte)
	AddHeaderFromPoolCalled       func(shardID uint32, nonce uint64)
}

// GetSelfHeaders -
func (bthm *BlockTrackerHandlerMock) GetSelfHeaders(headerHandler data.HeaderHandler) []*track.HeaderInfo {
	if bthm.GetSelfHeadersCalled != nil {
		return bthm.GetSelfHeadersCalled(headerHandler)
	}

	return nil
}

// ComputeCrossInfo -
func (bthm *BlockTrackerHandlerMock) ComputeCrossInfo(headers []data.HeaderHandler) {
	if bthm.ComputeCrossInfoCalled != nil {
		bthm.ComputeCrossInfoCalled(headers)
	}
}

// ComputeLongestSelfChain -
func (bthm *BlockTrackerHandlerMock) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	if bthm.ComputeLongestSelfChainCalled != nil {
		return bthm.ComputeLongestSelfChainCalled()
	}

	return nil, nil, nil, nil
}

// SortHeadersFromNonce -
func (bthm *BlockTrackerHandlerMock) SortHeadersFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	if bthm.SortHeadersFromNonceCalled != nil {
		return bthm.SortHeadersFromNonceCalled(shardID, nonce)
	}

	return nil, nil
}

// AddHeaderFromPool -
func (bthm *BlockTrackerHandlerMock) AddHeaderFromPool(shardID uint32, nonce uint64) {
	if bthm.AddHeaderFromPoolCalled != nil {
		bthm.AddHeaderFromPoolCalled(shardID, nonce)
	}
}

// IsInterfaceNil -
func (bthm *BlockTrackerHandlerMock) IsInterfaceNil() bool {
	return bthm == nil
}
