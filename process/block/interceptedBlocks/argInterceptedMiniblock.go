package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgInterceptedMiniblock is the argument for the intercepted miniblock
type ArgInterceptedMiniblock struct {
	MiniblockBuff    []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
