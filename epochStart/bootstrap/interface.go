package bootstrap

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	storageRequestFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart"
	bootStrapFactory "github.com/multiversx/mx-chain-go/epochStart/bootstrap/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	syncerFactory "github.com/multiversx/mx-chain-go/state/syncer/factory"
	updateSync "github.com/multiversx/mx-chain-go/update/sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// StartOfEpochNodesConfigHandler defines the methods to process nodesConfig from epoch start metablocks
type StartOfEpochNodesConfigHandler interface {
	NodesConfigFromMetaBlock(currMetaBlock data.HeaderHandler, prevMetaBlock data.HeaderHandler) (nodesCoordinator.NodesCoordinatorRegistryHandler, uint32, []*block.MiniBlock, error)
	IsInterfaceNil() bool
}

// EpochStartMetaBlockInterceptorProcessor defines the methods to sync an epoch start metablock
type EpochStartMetaBlockInterceptorProcessor interface {
	process.InterceptorProcessor
	GetEpochStartMetaBlock(ctx context.Context) (data.MetaHeaderHandler, error)
}

// StartInEpochNodesCoordinator defines the methods to process and save nodesCoordinator information to storage
type StartInEpochNodesCoordinator interface {
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NodesCoordinatorToRegistry(epoch uint32) nodesCoordinator.NodesCoordinatorRegistryHandler
	ShardIdForEpoch(epoch uint32) (uint32, error)
	IsInterfaceNil() bool
}

// Messenger defines which methods a p2p messenger should implement
type Messenger interface {
	dataRetriever.MessageHandler
	dataRetriever.TopicHandler
	UnregisterMessageProcessor(topic string, identifier string) error
	UnregisterAllMessageProcessors() error
	UnJoinAllTopics() error
	ConnectedPeers() []core.PeerID
	Verify(payload []byte, pid core.PeerID, signature []byte) error
	Broadcast(topic string, buff []byte)
	BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	Sign(payload []byte) ([]byte, error)
	SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error)
}

// RequestHandler defines which methods a request handler should implement
type RequestHandler interface {
	RequestStartOfEpochMetaBlock(epoch uint32)
	SetNumPeersToQuery(topic string, intra int, cross int) error
	GetNumPeersToQuery(topic string) (int, int, error)
	IsInterfaceNil() bool
}

// NodeTypeProviderHandler defines the actions needed for a component that can handle the node type
type NodeTypeProviderHandler interface {
	SetType(nodeType core.NodeType)
	GetType() core.NodeType
	IsInterfaceNil() bool
}

// EpochStartBootstrapperCreator defines the epoch start bootstrapper factory handler
type EpochStartBootstrapperCreator interface {
	CreateEpochStartBootstrapper(epochStartBootstrapArgs ArgsEpochStartBootstrap) (EpochStartBootstrapper, error)
	CreateStorageEpochStartBootstrapper(epochStartBootstrapArgs ArgsStorageEpochStartBootstrap) (EpochStartBootstrapper, error)
	IsInterfaceNil() bool
}

// EpochStartBootstrapper defines the epoch start bootstrap functionality
type EpochStartBootstrapper interface {
	Bootstrap() (Parameters, error)
	Close() error
	IsInterfaceNil() bool
}

type RunTypeComponentsHolder interface {
	AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator
	ShardCoordinatorCreator() sharding.ShardCoordinatorFactory
	NodesCoordinatorWithRaterCreator() nodesCoordinator.NodesCoordinatorWithRaterFactory
	RequestHandlerCreator() requestHandlers.RequestHandlerCreator
	RequestersContainerFactoryCreator() requesterscontainer.RequesterContainerFactoryCreator
	ValidatorAccountsSyncerFactoryHandler() syncerFactory.ValidatorAccountsSyncerFactoryHandler
	ShardRequestersContainerCreatorHandler() storageRequestFactory.ShardRequestersContainerCreatorHandler
	OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool
	DataCodecHandler() sovereign.DataCodecHandler
	TopicsCheckerHandler() sovereign.TopicsCheckerHandler
	AccountsCreator() state.AccountFactory
	IsInterfaceNil() bool
}

// ShardForLatestEpochComputer defines the methods to get the shard ID for the latest epoch
type ShardForLatestEpochComputer interface {
	GetShardIDForLatestEpoch() (uint32, bool, error)
	IsInterfaceNil() bool
}

type bootStrapShardProcessorHandler interface {
	requestAndProcessForShard(peerMiniBlocks []*block.MiniBlock) error
	computeNumShards(epochStartMeta data.MetaHeaderHandler) uint32
	createRequestHandler() (process.RequestHandler, error)
	createResolversContainer() error
	syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error)
	syncHeadersFromStorage(
		meta data.MetaHeaderHandler,
		syncingShardID uint32,
		importDBTargetShardID uint32,
		timeToWaitForRequestedData time.Duration,
	) (map[string]data.HeaderHandler, error)
	processNodesConfigFromStorage(pubKey []byte, importDBTargetShardID uint32) (nodesCoordinator.NodesCoordinatorRegistryHandler, uint32, error)
	createEpochStartMetaSyncer() (epochStart.StartOfEpochMetaSyncer, error)
	createStorageEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (epochStart.StartOfEpochMetaSyncer, error)
	createEpochStartInterceptorsContainers(args bootStrapFactory.ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, process.InterceptorsContainer, error)
	createCrossHeaderRequester() (updateSync.CrossHeaderRequester, error)
}

type epochStartTopicProviderHandler interface {
	getTopic() string
}

type epochStartPeerHandler interface {
	epochStartTopicProviderHandler
	setNumPeers(requestHandler RequestHandler, intra int, cross int) error
}

type shardTriggerRegistryHandler interface {
	GetEpochStartHeaderHandler() data.HeaderHandler
}
