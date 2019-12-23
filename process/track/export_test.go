package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

func (bbt *baseBlockTrack) LastHeaderForShard(shardID uint32) data.HeaderHandler {
	return bbt.lastHeaderForShard(shardID)
}

func (bbt *baseBlockTrack) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	bbt.addHeader(header, hash)
}
