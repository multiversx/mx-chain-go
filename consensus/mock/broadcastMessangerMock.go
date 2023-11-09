package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// BroadcastMessengerMock -
type BroadcastMessengerMock struct {
	BroadcastBlockCalled                     func(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeaderCalled                    func(data.HeaderHandler, []byte) error
	PrepareBroadcastBlockDataValidatorCalled func(h data.HeaderHandler, mbs map[uint32][]byte, txs map[string][][]byte, idx int, pkBytes []byte) error
	PrepareBroadcastHeaderValidatorCalled    func(h data.HeaderHandler, mbs map[uint32][]byte, txs map[string][][]byte, idx int, pkBytes []byte)
	BroadcastMiniBlocksCalled                func(map[uint32][]byte, []byte) error
	BroadcastTransactionsCalled              func(map[string][][]byte, []byte) error
	BroadcastConsensusMessageCalled          func(*consensus.Message) error
	BroadcastBlockDataLeaderCalled           func(h data.HeaderHandler, mbs map[uint32][]byte, txs map[string][][]byte, pkBytes []byte) error
}

// BroadcastBlock -
func (bmm *BroadcastMessengerMock) BroadcastBlock(bodyHandler data.BodyHandler, headerhandler data.HeaderHandler) error {
	if bmm.BroadcastBlockCalled != nil {
		return bmm.BroadcastBlockCalled(bodyHandler, headerhandler)
	}
	return nil
}

// BroadcastMiniBlocks -
func (bmm *BroadcastMessengerMock) BroadcastMiniBlocks(miniBlocks map[uint32][]byte, pkBytes []byte) error {
	if bmm.BroadcastMiniBlocksCalled != nil {
		return bmm.BroadcastMiniBlocksCalled(miniBlocks, pkBytes)
	}
	return nil
}

// BroadcastBlockDataLeader -
func (bmm *BroadcastMessengerMock) BroadcastBlockDataLeader(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	pkBytes []byte,
) error {
	if bmm.BroadcastBlockDataLeaderCalled != nil {
		return bmm.BroadcastBlockDataLeaderCalled(header, miniBlocks, transactions, pkBytes)
	}

	err := bmm.BroadcastMiniBlocks(miniBlocks, pkBytes)
	if err != nil {
		return err
	}

	return bmm.BroadcastTransactions(transactions, pkBytes)
}

// PrepareBroadcastBlockDataValidator -
func (bmm *BroadcastMessengerMock) PrepareBroadcastBlockDataValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	idx int,
	pkBytes []byte,
) {
	if bmm.PrepareBroadcastBlockDataValidatorCalled != nil {
		_ = bmm.PrepareBroadcastBlockDataValidatorCalled(
			header,
			miniBlocks,
			transactions,
			idx,
			pkBytes,
		)
	}
}

// PrepareBroadcastHeaderValidator -
func (bmm *BroadcastMessengerMock) PrepareBroadcastHeaderValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	order int,
	pkBytes []byte,
) {
	if bmm.PrepareBroadcastHeaderValidatorCalled != nil {
		bmm.PrepareBroadcastHeaderValidatorCalled(
			header,
			miniBlocks,
			transactions,
			order,
			pkBytes,
		)
	}
}

// BroadcastTransactions -
func (bmm *BroadcastMessengerMock) BroadcastTransactions(transactions map[string][][]byte, pkBytes []byte) error {
	if bmm.BroadcastTransactionsCalled != nil {
		return bmm.BroadcastTransactionsCalled(transactions, pkBytes)
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
func (bmm *BroadcastMessengerMock) BroadcastHeader(headerhandler data.HeaderHandler, pkBytes []byte) error {
	if bmm.BroadcastHeaderCalled != nil {
		return bmm.BroadcastHeaderCalled(headerhandler, pkBytes)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bmm *BroadcastMessengerMock) IsInterfaceNil() bool {
	return bmm == nil
}
