package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockTrackerStub struct {
	IsShardStuckCalled                 func(shardId uint32) bool
	LastHeaderForShardCalled           func(shardId uint32) data.HeaderHandler
	RegisterSelfNotarizedHandlerCalled func(handler func(headers []data.HeaderHandler))
}

func (bts *BlockTrackerStub) IsShardStuck(shardId uint32) bool {
	return bts.IsShardStuckCalled(shardId)
}

func (bts *BlockTrackerStub) LastHeaderForShard(shardId uint32) data.HeaderHandler {
	return bts.LastHeaderForShardCalled(shardId)
}

func (bts *BlockTrackerStub) RegisterSelfNotarizedHandler(handler func(headers []data.HeaderHandler)) {
	bts.RegisterSelfNotarizedHandlerCalled(handler)
}

func (bts *BlockTrackerStub) IsInterfaceNil() bool {
	return bts == nil
}
