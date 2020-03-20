package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// PendingMiniBlocksHandlerStub -
type PendingMiniBlocksHandlerStub struct {
	AddProcessedHeaderCalled   func(handler data.HeaderHandler) error
	RevertHeaderCalled         func(handler data.HeaderHandler) error
	GetPendingMiniBlocksCalled func(shardID uint32) [][]byte
	SetPendingMiniBlocksCalled func(shardID uint32, mbHashes [][]byte)
}

// AddProcessedHeader -
func (p *PendingMiniBlocksHandlerStub) AddProcessedHeader(handler data.HeaderHandler) error {
	if p.AddProcessedHeaderCalled != nil {
		return p.AddProcessedHeaderCalled(handler)
	}
	return nil
}

// RevertHeader -
func (p *PendingMiniBlocksHandlerStub) RevertHeader(handler data.HeaderHandler) error {
	if p.RevertHeaderCalled != nil {
		return p.RevertHeaderCalled(handler)
	}
	return nil
}

// GetPendingMiniBlocks -
func (p *PendingMiniBlocksHandlerStub) GetPendingMiniBlocks(shardID uint32) [][]byte {
	if p.GetPendingMiniBlocksCalled != nil {
		return p.GetPendingMiniBlocksCalled(shardID)
	}
	return make([][]byte, 0)
}

// SetPendingMiniBlocks -
func (p *PendingMiniBlocksHandlerStub) SetPendingMiniBlocks(shardID uint32, mbHashes [][]byte) {
	if p.SetPendingMiniBlocksCalled != nil {
		p.SetPendingMiniBlocksCalled(shardID, mbHashes)
	}
}

// IsInterfaceNil -
func (p *PendingMiniBlocksHandlerStub) IsInterfaceNil() bool {
	return p == nil
}
