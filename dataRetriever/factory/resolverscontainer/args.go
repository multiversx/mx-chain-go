package resolverscontainer

import (
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
	TriesContainer              common.TriesHolder
	InputAntifloodHandler       dataRetriever.P2PAntifloodHandler
	OutputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	CurrentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	PreferredPeersHolder        p2p.PreferredPeersHolderHandler
	PeersRatingHandler          dataRetriever.PeersRatingHandler
	SizeCheckDelta              uint32
	IsFullHistoryNode           bool
	NodesCoordinator                     dataRetriever.NodesCoordinator
	MaxNumOfPeerAuthenticationInResponse int
	PeerShardMapper                      process.PeerShardMapper
}
