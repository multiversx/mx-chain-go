package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlocksTrackerMock struct {
	UnnotarisedBlocksCalled      func() []data.HeaderHandler
	RemoveNotarisedBlocksCalled  func(headerHandler data.HeaderHandler) error
	AddBlockCalled               func(headerHandler data.HeaderHandler)
	SetBlockBroadcastRoundCalled func(nonce uint64, round int32)
	BlockBroadcastRoundCalled    func(nonce uint64) int32
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

func (btm *BlocksTrackerMock) SetBlockBroadcastRound(nonce uint64, round int32) {
	btm.SetBlockBroadcastRoundCalled(nonce, round)
}

func (btm *BlocksTrackerMock) BlockBroadcastRound(nonce uint64) int32 {
	return btm.BlockBroadcastRoundCalled(nonce)
}
