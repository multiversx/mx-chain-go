package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgInterceptedBlockHeader is the argument for the intercepted header
type ArgInterceptedBlockHeader struct {
	HdrBuff          []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	MultiSigVerifier crypto.MultiSigVerifier
	NodesCoordinator sharding.NodesCoordinator
	ShardCoordinator sharding.Coordinator
}
