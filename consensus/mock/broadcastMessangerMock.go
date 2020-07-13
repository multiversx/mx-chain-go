package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
)

// BroadcastMessengerMock -
type BroadcastMessengerMock struct {
	BroadcastBlockCalled                     func(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeaderCalled                    func(data.HeaderHandler) error
	PrepareBroadcastBlockDataValidatorCalled func(h data.HeaderHandler, mbs map[uint32][]byte, txs map[string][][]byte, idx int) error
	PrepareBroadcastHeaderValidatorCalled    func(h data.HeaderHandler, mbs map[uint32][]byte, txs map[string][][]byte, idx int)
	BroadcastMiniBlocksCalled                func(map[uint32][]byte) error
	BroadcastTransactionsCalled              func(map[string][][]byte) error
	BroadcastConsensusMessageCalled          func(*consensus.Message) error
	BroadcastBlockDataLeaderCalled           func(h data.HeaderHandler, mbs map[uint32][]byte, txs map[string][][]byte) error
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

// BroadcastBlockDataLeader -
func (bmm *BroadcastMessengerMock) BroadcastBlockDataLeader(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) error {
	if bmm.BroadcastBlockDataLeaderCalled != nil {
		return bmm.BroadcastBlockDataLeaderCalled(header, miniBlocks, transactions)
	}

	err := bmm.BroadcastMiniBlocks(miniBlocks)
	if err != nil {
		return err
	}

	return bmm.BroadcastTransactions(transactions)
}

// PrepareBroadcastBlockDataValidator -
func (bmm *BroadcastMessengerMock) PrepareBroadcastBlockDataValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	idx int,
) {
	if bmm.PrepareBroadcastBlockDataValidatorCalled != nil {
		_ = bmm.PrepareBroadcastBlockDataValidatorCalled(
			header,
			miniBlocks,
			transactions,
			idx,
		)
	}
}

// PrepareBroadcastHeaderValidator -
func (bmm *BroadcastMessengerMock) PrepareBroadcastHeaderValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	order int,
) {
	if bmm.PrepareBroadcastHeaderValidatorCalled != nil {
		bmm.PrepareBroadcastHeaderValidatorCalled(
			header,
			miniBlocks,
			transactions,
			order,
		)
	}
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
