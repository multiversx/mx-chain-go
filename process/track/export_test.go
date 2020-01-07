package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

func (bbt *baseBlockTrack) GetLastHeader(shardID uint32) data.HeaderHandler {
	return bbt.getLastHeader(shardID)
}
