package factory

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgInterceptedDataFactory holds all dependencies required by the intercepted data factory in order to create
// new instances
type ArgInterceptedDataFactory struct {
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	KeyGen           crypto.KeyGenerator
	Signer           crypto.SingleSigner
	AddrConv         state.AddressConverter
	ShardCoordinator sharding.Coordinator
}
