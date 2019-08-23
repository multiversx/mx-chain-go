package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlocksTrackerMock struct {
	UnnotarisedBlocksCalled      func() []data.HeaderHandler
	RemoveNotarisedBlocksCalled  func(headerHandler data.HeaderHandler) error
	AddBlockCalled               func(headerHandler data.HeaderHandler)
	SetBlockBroadcastRoundCalled func(nonce uint64, round int64)
	BlockBroadcastRoundCalled    func(nonce uint64) int64
}

func (btm *BlocksTrackerMock) UnnotarisedBlocks() []data.HeaderHandler {
	return btm.UnnotarisedBlocksCalled()
}

func (btm *BlocksTrackerMock) RemoveNotarisedBlocks(headerHandler data.HeaderHandler) error {
	return btm.RemoveNotarisedBlocksCalled(headerHandler)
}

func (btm *BlocksTrackerMock) AddBlock(headerHandler data.HeaderHandler) {
	btm.AddBlockCalled(headerHandler)
}

func (btm *BlocksTrackerMock) SetBlockBroadcastRound(nonce uint64, round int64) {
	btm.SetBlockBroadcastRoundCalled(nonce, round)
}

func (btm *BlocksTrackerMock) BlockBroadcastRound(nonce uint64) int64 {
	return btm.BlockBroadcastRoundCalled(nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (btm *BlocksTrackerMock) IsInterfaceNil() bool {
	if btm == nil {
		return true
	}
	return false
}
