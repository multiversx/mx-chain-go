package requesterscontainer

import (
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
