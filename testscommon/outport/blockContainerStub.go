package outport

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// BlockContainerStub -
type BlockContainerStub struct {
	GetCalled func(headerType core.HeaderType) (block.EmptyBlockCreator, error)
}

// Get -
func (bcs *BlockContainerStub) Get(headerType core.HeaderType) (block.EmptyBlockCreator, error) {
	if bcs.GetCalled != nil {
		return bcs.GetCalled(headerType)
	}

	return nil, nil
}
