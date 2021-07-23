package resolverscontainer

import (
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

// FactoryArgs will hold the arguments for ResolversContainerFactory for both shard and meta
type FactoryArgs struct {
	ResolverConfig              config.ResolverConfig
	NumConcurrentResolvingJobs  int32
	ShardCoordinator            sharding.Coordinator
	Messenger                   dataRetriever.TopicMessageHandler
	Store                       dataRetriever.StorageService
	Marshalizer                 marshal.Marshalizer
	DataPools                   dataRetriever.PoolsHolder
	Uint64ByteSliceConverter    typeConverters.Uint64ByteSliceConverter
	DataPacker                  dataRetriever.DataPacker
	TriesContainer              state.TriesHolder
	InputAntifloodHandler       dataRetriever.P2PAntifloodHandler
	OutputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	CurrentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	PreferredPeersHolder        p2p.PreferredPeersHolderHandler
	SizeCheckDelta              uint32
	IsFullHistoryNode           bool
}
