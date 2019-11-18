package hooks

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgBlockChainHook represents the arguments structure for the blockchain hook
type ArgBlockChainHook struct {
	Accounts         state.AccountsAdapter
	AddrConv         state.AddressConverter
	StorageService   dataRetriever.StorageService
	BlockChain       data.ChainHandler
	ShardCoordinator sharding.Coordinator
	Marshalizer      marshal.Marshalizer
	Uint64Converter  typeConverters.Uint64ByteSliceConverter
}
