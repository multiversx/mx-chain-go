package processor

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgTxBodyInterceptorProcessor is the argument for the interceptor processor used for tx block body
type ArgTxBodyInterceptorProcessor struct {
	Miniblocks       storage.Cacher
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
