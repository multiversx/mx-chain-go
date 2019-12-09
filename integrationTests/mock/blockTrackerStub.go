package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	IsShardStuckCalled       func(shardId uint32) bool
	AddHeaderCalled          func(header data.HeaderHandler, hash []byte)
	LastHeaderForShardCalled func(shardId uint32) data.HeaderHandler
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	return bts.IsShardStuckCalled(shardId)
}

func (bts *BlockTrackerStub) AddHeader(header data.HeaderHandler, hash []byte) {
	bts.AddHeaderCalled(header, hash)
}

func (bts *BlockTrackerStub) LastHeaderForShard(shardId uint32) data.HeaderHandler {
	return bts.LastHeaderForShardCalled(shardId)
}

func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
