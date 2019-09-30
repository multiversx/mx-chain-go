package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgInterceptedTxBlockBody is the argument for the intercepted tx block body
type ArgInterceptedTxBlockBody struct {
	TxBlockBodyBuff  []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
