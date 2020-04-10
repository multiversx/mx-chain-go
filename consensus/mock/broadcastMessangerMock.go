package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
)

// BroadcastMessengerMock -
type BroadcastMessengerMock struct {
	BroadcastBlockCalled            func(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeaderCalled           func(data.HeaderHandler) error
	SetDataForDelayBroadcastCalled  func([]byte, map[uint32][]byte, map[string][][]byte) error
	BroadcastMiniBlocksCalled       func(map[uint32][]byte) error
	BroadcastTransactionsCalled     func(map[string][][]byte) error
	BroadcastConsensusMessageCalled func(*consensus.Message) error
}

// BroadcastBlock -
func (bmm *BroadcastMessengerMock) BroadcastBlock(bodyHandler data.BodyHandler, headerhandler data.HeaderHandler) error {
	if bmm.BroadcastBlockCalled != nil {
		return bmm.BroadcastBlockCalled(bodyHandler, headerhandler)
	}
	return nil
}

// BroadcastMiniBlocks -
func (bmm *BroadcastMessengerMock) BroadcastMiniBlocks(miniBlocks map[uint32][]byte) error {
	if bmm.BroadcastMiniBlocksCalled != nil {
		return bmm.BroadcastMiniBlocksCalled(miniBlocks)
	}
	return nil
}

// SetDataForDelayBroadcast -
func (bmm *BroadcastMessengerMock) SetDataForDelayBroadcast(
	headerHash []byte,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) error {
	if bmm.SetDataForDelayBroadcastCalled != nil {
		return bmm.SetDataForDelayBroadcastCalled(headerHash, miniBlocks, transactions)
	}
	return nil
}

// BroadcastTransactions -
func (bmm *BroadcastMessengerMock) BroadcastTransactions(transactions map[string][][]byte) error {
	if bmm.BroadcastTransactionsCalled != nil {
		return bmm.BroadcastTransactionsCalled(transactions)
	}
	return nil
}

// BroadcastConsensusMessage -
func (bmm *BroadcastMessengerMock) BroadcastConsensusMessage(message *consensus.Message) error {
	if bmm.BroadcastConsensusMessageCalled != nil {
		return bmm.BroadcastConsensusMessageCalled(message)
	}
	return nil
}

// BroadcastHeader -
func (bmm *BroadcastMessengerMock) BroadcastHeader(headerhandler data.HeaderHandler) error {
	if bmm.BroadcastHeaderCalled != nil {
		return bmm.BroadcastHeaderCalled(headerhandler)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bmm *BroadcastMessengerMock) IsInterfaceNil() bool {
	return bmm == nil
}
