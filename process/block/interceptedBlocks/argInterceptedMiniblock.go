package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgInterceptedMiniblock is the argument for the intercepted miniblock
type ArgInterceptedMiniblock struct {
	MiniblockBuff    []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
