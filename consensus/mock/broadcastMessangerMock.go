package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type BroadcastMessangerMock struct {
	BroadcastBlockCalled                     func(data.BodyHandler, data.HeaderHandler) error
	BroadcastConsensusMessageCalled          func(*consensus.Message) error
	BroadcastHeaderCalled                    func(data.HeaderHandler) error
	BroadcastMiniBlocksAndTransactionsCalled func(data.BodyHandler, data.HeaderHandler) error
}

func (bmm *BroadcastMessangerMock) BroadcastBlock() func(data.BodyHandler, data.HeaderHandler) error {
	if bmm.BroadcastBlockCalled != nil {
		return bmm.BroadcastBlockCalled
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastConsensusMessage() func(*consensus.Message) error {
	if bmm.BroadcastConsensusMessageCalled != nil {
		return bmm.BroadcastConsensusMessageCalled
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastHeader() func(data.HeaderHandler) error {
	if bmm.BroadcastHeaderCalled != nil {
		return bmm.BroadcastHeaderCalled
	}
	return nil
}

func (bmm *BroadcastMessangerMock) BroadcastMiniBlocksAndTransactions() func(data.BodyHandler, data.HeaderHandler) error {
	if bmm.BroadcastMiniBlocksAndTransactionsCalled != nil {
		return bmm.BroadcastMiniBlocksAndTransactionsCalled
	}
	return nil
}
