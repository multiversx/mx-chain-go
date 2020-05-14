package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// StartOfEpochNodesConfigHandler defines the methods to process nodesConfig from epoch start metablocks
type StartOfEpochNodesConfigHandler interface {
	NodesConfigFromMetaBlock(currMetaBlock *block.MetaBlock, prevMetaBlock *block.MetaBlock) (*sharding.NodesCoordinatorRegistry, uint32, error)
	IsInterfaceNil() bool
}

// EpochStartInterceptor defines the methods to sync an epoch start metablock
type EpochStartInterceptor interface {
	process.Interceptor
	GetEpochStartMetaBlock(target int, epoch uint32) (*block.MetaBlock, error)
}

// StartInEpochNodesCoordinator defines the methods to process and save nodesCoordinator information to storage
type StartInEpochNodesCoordinator interface {
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NodesCoordinatorToRegistry() *sharding.NodesCoordinatorRegistry
	ShardIdForEpoch(epoch uint32) (uint32, error)
	IsInterfaceNil() bool
}

// Messenger defines which methods a p2p messenger should implement
type Messenger interface {
	dataRetriever.MessageHandler
	dataRetriever.TopicHandler
	UnregisterMessageProcessor(topic string) error
	UnregisterAllMessageProcessors() error
	ConnectedPeers() []p2p.PeerID
}

// RequestHandler defines which methods a request handler should implement
type RequestHandler interface {
	RequestStartOfEpochMetaBlock(epoch uint32)
	SetNumPeersToQuery(topic string, intra int, cross int) error
	GetNumPeersToQuery(topic string) (int, int, error)
	IsInterfaceNil() bool
}

// BootstrapCryptoComponentsHolder holds the crypto components required for bootstrap
type BootstrapCryptoComponentsHolder interface {
	PublicKey() crypto.PublicKey
	BlockSigner() crypto.SingleSigner
	TxSingleSigner() crypto.SingleSigner
	MultiSigner() crypto.MultiSigner
	BlockSignKeyGen() crypto.KeyGenerator
	TxSignKeyGen() crypto.KeyGenerator
	IsInterfaceNil() bool
}

// BootstrapCoreComponentsHolder holds the core components required for bootstrap
type BootstrapCoreComponentsHolder interface {
	InternalMarshalizer() marshal.Marshalizer
	TxMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverter() state.PubkeyConverter
	PathHandler() storage.PathManagerHandler
	ChainID() string
	IsInterfaceNil() bool
}
