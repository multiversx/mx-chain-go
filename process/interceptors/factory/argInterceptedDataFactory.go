package factory

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgShardInterceptedDataFactory holds all dependencies required by the shard intercepted data factory in order to create
// new instances
type ArgShardInterceptedDataFactory struct {
	*ArgMetaInterceptedDataFactory
	KeyGen     crypto.KeyGenerator
	Signer     crypto.SingleSigner
	AddrConv   state.AddressConverter
	FeeHandler process.FeeHandler
}

// ArgMetaInterceptedDataFactory holds all dependencies required by the meta intercepted data factory in order to create
// new instances
type ArgMetaInterceptedDataFactory struct {
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
	MultiSigVerifier crypto.MultiSigVerifier
	NodesCoordinator sharding.NodesCoordinator
}
