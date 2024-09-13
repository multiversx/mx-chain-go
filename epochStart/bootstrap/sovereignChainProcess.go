package bootstrap

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/heartbeat/validator"
	"github.com/multiversx/mx-chain-go/storage/cache"
)

type sovereignChainEpochStartBootstrap struct {
	*epochStartBootstrap
}

// NewSovereignChainEpochStartBootstrap creates a new instance of sovereignChainEpochStartBootstrap
func NewSovereignChainEpochStartBootstrap(epochStartBootstrap *epochStartBootstrap) (*sovereignChainEpochStartBootstrap, error) {
	if epochStartBootstrap == nil {
		return nil, errors.ErrNilEpochStartBootstrapper
	}

	scesb := &sovereignChainEpochStartBootstrap{
		epochStartBootstrap,
	}

	scesb.bootStrapShardRequester = &sovereignBootStrapShardRequester{
		scesb,
	}

	scesb.getDataToSyncMethod = scesb.getDataToSync
	scesb.shardForLatestEpochComputer = scesb
	return scesb, nil
}

// todo : probably delete this
func (scesb *sovereignChainEpochStartBootstrap) getDataToSync(
	_ data.EpochStartShardDataHandler,
	shardNotarizedHeader data.ShardHeaderHandler,
) (*dataToSync, error) {
	return &dataToSync{
		ownShardHdr:       shardNotarizedHeader,
		rootHashToSync:    shardNotarizedHeader.GetRootHash(),
		withScheduled:     false,
		additionalHeaders: nil,
	}, nil
}

// GetShardIDForLatestEpoch returns the shard ID for the latest epoch
func (scesb *sovereignChainEpochStartBootstrap) GetShardIDForLatestEpoch() (uint32, bool, error) {
	return core.SovereignChainShardId, false, nil
}

func (e *sovereignChainEpochStartBootstrap) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, 1)
	shardIds := make([]uint32, 0, 1)

	if meta.GetEpoch() > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.SovereignChainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == e.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.SovereignChainHeader{}
	}

	return syncedHeaders, nil
}

func (e *sovereignChainEpochStartBootstrap) createResolversContainer() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(e.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	storageService := disabled.NewChainStorer()

	payloadValidator, err := validator.NewPeerAuthenticationPayloadValidator(e.generalConfig.HeartbeatV2.HeartbeatExpiryTimespanInSec)
	if err != nil {
		return err
	}

	// TODO - create a dedicated request handler to be used when fetching required data with the correct shard coordinator
	//  this one should only be used before determining the correct shard where the node should reside
	log.Debug("epochStartBootstrap.createRequestHandler", "shard", e.shardCoordinator.SelfId())
	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:                    e.shardCoordinator,
		MainMessenger:                       e.mainMessenger,
		FullArchiveMessenger:                e.fullArchiveMessenger,
		Store:                               storageService,
		Marshalizer:                         e.coreComponentsHolder.InternalMarshalizer(),
		DataPools:                           e.dataPool,
		Uint64ByteSliceConverter:            uint64ByteSlice.NewBigEndianConverter(),
		NumConcurrentResolvingJobs:          10,
		NumConcurrentResolvingTrieNodesJobs: 3,
		DataPacker:                          dataPacker,
		TriesContainer:                      e.trieContainer,
		SizeCheckDelta:                      0,
		InputAntifloodHandler:               disabled.NewAntiFloodHandler(),
		OutputAntifloodHandler:              disabled.NewAntiFloodHandler(),
		MainPreferredPeersHolder:            disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder:     disabled.NewPreferredPeersHolder(),
		PayloadValidator:                    payloadValidator,
	}

	sp, err := resolverscontainer.NewShardResolversContainerFactory(resolversContainerArgs)

	resolverFactory, err := resolverscontainer.NewSovereignShardResolversContainerFactory(sp)
	if err != nil {
		return err
	}

	container, err := resolverFactory.Create()
	if err != nil {
		return err
	}

	_ = container
	return nil
}

func (e *sovereignChainEpochStartBootstrap) createRequestHandler() error {
	requestersContainerArgs := requesterscontainer.FactoryArgs{
		RequesterConfig:                 e.generalConfig.Requesters,
		ShardCoordinator:                e.shardCoordinator,
		MainMessenger:                   e.mainMessenger,
		FullArchiveMessenger:            e.fullArchiveMessenger,
		Marshaller:                      e.coreComponentsHolder.InternalMarshalizer(),
		Uint64ByteSliceConverter:        uint64ByteSlice.NewBigEndianConverter(),
		OutputAntifloodHandler:          disabled.NewAntiFloodHandler(),
		CurrentNetworkEpochProvider:     disabled.NewCurrentNetworkEpochProviderHandler(),
		MainPreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		PeersRatingHandler:              disabled.NewDisabledPeersRatingHandler(),
		SizeCheckDelta:                  0,
	}

	sh, err := requesterscontainer.NewShardRequestersContainerFactory(requestersContainerArgs)
	if err != nil {
		return err
	}
	requestersFactory, err := requesterscontainer.NewSovereignShardRequestersContainerFactory(sh)
	if err != nil {
		return err
	}

	container, err := requestersFactory.Create()
	if err != nil {
		return err
	}

	finder, err := containers.NewRequestersFinder(container, e.shardCoordinator)
	if err != nil {
		return err
	}

	requestedItemsHandler := cache.NewTimeCache(timeBetweenRequests)
	e.requestHandler, err = e.runTypeComponents.RequestHandlerCreator().CreateRequestHandler(
		requestHandlers.RequestHandlerArgs{
			RequestersFinder:      finder,
			RequestedItemsHandler: requestedItemsHandler,
			WhiteListHandler:      e.whiteListHandler,
			MaxTxsToRequest:       maxToRequest,
			ShardID:               core.MetachainShardId,
			RequestInterval:       timeBetweenRequests,
		},
	)
	return err
}

func (e *sovereignChainEpochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
