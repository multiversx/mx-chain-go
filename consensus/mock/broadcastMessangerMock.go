package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type BroadcastMessangerMock struct {
	BroadcastBlockCalled            func(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeaderCalled           func(data.HeaderHandler) error
	BroadcastMiniBlocksCalled       func(map[uint32][]byte) error
	BroadcastTransactionsCalled     func(map[uint32][][]byte) error
	BroadcastConsensusMessageCalled func(*consensus.Message) error
}

func (bmm *BroadcastMessangerMock) BroadcastBlock(bodyHandler data.BodyHandler, headerhandler data.HeaderHandler) error {
	if bmm.BroadcastBlockCalled != nil {
		return bmm.BroadcastBlockCalled(bodyHandler, headerhandler)
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastHeader(headerHandler data.HeaderHandler) error {
	if bmm.BroadcastHeaderCalled != nil {
		return bmm.BroadcastHeaderCalled(headerHandler)
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastMiniBlocks(miniBlocks map[uint32][]byte) error {
	if bmm.BroadcastMiniBlocksCalled != nil {
		return bmm.BroadcastMiniBlocksCalled(miniBlocks)
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastTransactions(transactions map[uint32][][]byte) error {
	if bmm.BroadcastTransactionsCalled != nil {
		return bmm.BroadcastTransactionsCalled(transactions)
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastConsensusMessage(message *consensus.Message) error {
	if bmm.BroadcastConsensusMessageCalled != nil {
		return bmm.BroadcastConsensusMessageCalled(message)
	}
	return nil
}
