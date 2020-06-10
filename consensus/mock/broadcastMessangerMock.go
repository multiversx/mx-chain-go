package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
)

// BroadcastMessengerMock -
type BroadcastMessengerMock struct {
	BroadcastBlockCalled             func(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeaderCalled            func(data.HeaderHandler) error
	SetDataForDelayBroadcastCalled   func([]byte, map[uint32][]byte, map[string][][]byte) error
	SetValidatorDelayBroadcastCalled func(
		headerHash []byte,
		prevRandSeed []byte,
		round uint64,
		miniBlocks map[uint32][]byte,
		miniBlockHashes map[uint32]map[string]struct{},
		transactions map[string][][]byte,
		order uint32,
	) error
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

// BroadcastBlockDataLeader -
func (bmm *BroadcastMessengerMock) BroadcastBlockDataLeader(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) error {
	return nil
}

// BroadcastBlockDataValidator -
func (bmm *BroadcastMessengerMock) PrepareBroadcastBlockDataValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	idx int,
) error {
	return nil
}

// SetLeaderDelayBroadcast -
func (bmm *BroadcastMessengerMock) SetLeaderDelayBroadcast(
	headerHash []byte,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) error {
	if bmm.SetDataForDelayBroadcastCalled != nil {
		return bmm.SetDataForDelayBroadcastCalled(headerHash, miniBlocks, transactions)
	}

	err := bmm.BroadcastMiniBlocks(miniBlocks)
	if err != nil {
		return err
	}

	return bmm.BroadcastTransactions(transactions)
}

// SetValidatorDelayBroadcast -
func (bmm *BroadcastMessengerMock) SetValidatorDelayBroadcast(headerHash []byte, prevRandSeed []byte, round uint64, miniBlocks map[uint32][]byte, miniBlockHashes map[uint32]map[string]struct{}, transactions map[string][][]byte, order uint32) error {
	if bmm.SetValidatorDelayBroadcastCalled != nil {
		return bmm.SetValidatorDelayBroadcastCalled(
			headerHash,
			prevRandSeed,
			round,
			miniBlocks,
			miniBlockHashes,
			transactions,
			order,
		)
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
