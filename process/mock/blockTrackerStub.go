package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	IsShardStuckCalled       func(shardId uint32) bool
	AddHeaderCalled          func(header data.HeaderHandler)
	LastHeaderForShardCalled func(shardId uint32) data.HeaderHandler
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	return bts.IsShardStuckCalled(shardId)
}

func (bts *BlockTrackerStub) AddHeader(header data.HeaderHandler) {
	bts.AddHeaderCalled(header)
}

func (bts *BlockTrackerStub) LastHeaderForShard(shardId uint32) data.HeaderHandler {
	return bts.LastHeaderForShardCalled(shardId)
}

func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
