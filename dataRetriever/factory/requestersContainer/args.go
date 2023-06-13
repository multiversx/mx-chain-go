package requesterscontainer

import (
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding"
)

// FactoryArgs will hold the arguments for RequestersContainerFactory for both shard and meta
type FactoryArgs struct {
	RequesterConfig             config.RequesterConfig
	ShardCoordinator            sharding.Coordinator
	Messenger                   dataRetriever.TopicMessageHandler
	Marshaller                  marshal.Marshalizer
	Uint64ByteSliceConverter    typeConverters.Uint64ByteSliceConverter
	OutputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	CurrentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	PreferredPeersHolder        p2p.PreferredPeersHolderHandler
	PeersRatingHandler          dataRetriever.PeersRatingHandler
	SizeCheckDelta              uint32
}
