package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
)

// BlockTrackerFactoryStub -
type BlockTrackerFactoryStub struct {
	CreateBlockTrackerCalled func(args track.ArgShardTracker) (process.BlockTracker, error)
}

// NewBlockTrackerFactoryStub -
func NewBlockTrackerFactoryStub() *BlockTrackerFactoryStub {
	return &BlockTrackerFactoryStub{}
}

// CreateBlockTracker -
func (b *BlockTrackerFactoryStub) CreateBlockTracker(args track.ArgShardTracker) (process.BlockTracker, error) {
	if b.CreateBlockTrackerCalled != nil {
		return b.CreateBlockTrackerCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BlockTrackerFactoryStub) IsInterfaceNil() bool {
	return false
}
