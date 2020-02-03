package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type PendingMiniBlocksHandlerStub struct {
	PendingMiniBlockHeadersCalled func(lastNotarizedHeaders []data.HeaderHandler) ([]block.ShardMiniBlockHeader, error)
	AddProcessedHeaderCalled      func(handler data.HeaderHandler) error
	RevertHeaderCalled            func(handler data.HeaderHandler) error
	GetNumPendingMiniBlocksCalled func(shardID uint32) uint32
	SetNumPendingMiniBlocksCalled func(shardID uint32, numPendingMiniBlocks uint32)
}

func (p *PendingMiniBlocksHandlerStub) PendingMiniBlockHeaders(lastNotarizedHeaders []data.HeaderHandler) ([]block.ShardMiniBlockHeader, error) {
	if p.PendingMiniBlockHeadersCalled != nil {
		return p.PendingMiniBlockHeadersCalled(lastNotarizedHeaders)
	}
	return nil, nil
}

func (p *PendingMiniBlocksHandlerStub) AddProcessedHeader(handler data.HeaderHandler) error {
	if p.AddProcessedHeaderCalled != nil {
		return p.AddProcessedHeaderCalled(handler)
	}
	return nil
}

func (p *PendingMiniBlocksHandlerStub) RevertHeader(handler data.HeaderHandler) error {
	if p.RevertHeaderCalled != nil {
		return p.RevertHeaderCalled(handler)
	}
	return nil
}

func (p *PendingMiniBlocksHandlerStub) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	if p.GetNumPendingMiniBlocksCalled != nil {
		return p.GetNumPendingMiniBlocksCalled(shardID)
	}
	return 0
}

func (p *PendingMiniBlocksHandlerStub) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	if p.SetNumPendingMiniBlocksCalled != nil {
		p.SetNumPendingMiniBlocksCalled(shardID, numPendingMiniBlocks)
	}
}

func (p *PendingMiniBlocksHandlerStub) IsInterfaceNil() bool {
	return p == nil
}
