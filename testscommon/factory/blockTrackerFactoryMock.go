package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BlockTrackerFactoryMock -
type BlockTrackerFactoryMock struct {
	CreateBlockTrackerCalled func(args track.ArgShardTracker) (process.BlockTracker, error)
}

// CreateBlockTracker -
func (b *BlockTrackerFactoryMock) CreateBlockTracker(args track.ArgShardTracker) (process.BlockTracker, error) {
	if b.CreateBlockTrackerCalled != nil {
		return b.CreateBlockTrackerCalled(args)
	}
	return &testscommon.BlockTrackerStub{}, nil
}

// IsInterfaceNil -
func (b *BlockTrackerFactoryMock) IsInterfaceNil() bool {
	return b == nil
}
