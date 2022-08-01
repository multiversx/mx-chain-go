package bootstrap

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/update/sync"
)

const consensusGroupCacheSize = 50

type syncValidatorStatus struct {
	miniBlocksSyncer   epochStart.PendingMiniBlocksSyncHandler
	dataPool           dataRetriever.PoolsHolder
	marshalizer        marshal.Marshalizer
	requestHandler     process.RequestHandler
	nodeCoordinator    StartInEpochNodesCoordinator
	genesisNodesConfig sharding.GenesisNodesSetupHandler
	memDB              storage.Storer
}

// ArgsNewSyncValidatorStatus holds the arguments needed for creating a new validator status process component
type ArgsNewSyncValidatorStatus struct {
	DataPool                  dataRetriever.PoolsHolder
	Marshalizer               marshal.Marshalizer
	Hasher                    hashing.Hasher
	RequestHandler            process.RequestHandler
	ChanceComputer            nodesCoordinator.ChanceComputer
	GenesisNodesConfig        sharding.GenesisNodesSetupHandler
	NodeShuffler              nodesCoordinator.NodesShuffler
	PubKey                    []byte
	ShardIdAsObserver         uint32
	WaitingListFixEnableEpoch uint32
	ChanNodeStop              chan endProcess.ArgEndProcess
	NodeTypeProvider          NodeTypeProviderHandler
	IsFullArchive             bool
}

// NewSyncValidatorStatus creates a new validator status process component
func NewSyncValidatorStatus(args ArgsNewSyncValidatorStatus) (*syncValidatorStatus, error) {
	if args.ChanNodeStop == nil {
		return nil, nodesCoordinator.ErrNilNodeStopChannel
	}

	s := &syncValidatorStatus{
		dataPool:           args.DataPool,
		marshalizer:        args.Marshalizer,
		requestHandler:     args.RequestHandler,
		genesisNodesConfig: args.GenesisNodesConfig,
	}
	syncMiniBlocksArgs := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          s.dataPool.MiniBlocks(),
		Marshalizer:    s.marshalizer,
		RequestHandler: s.requestHandler,
	}
	var err error
	s.miniBlocksSyncer, err = sync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
	if err != nil {
		return nil, err
	}

	eligibleNodesInfo, waitingNodesInfo := args.GenesisNodesConfig.InitialNodesInfo()

	eligibleValidators, err := nodesCoordinator.NodesInfoToValidators(eligibleNodesInfo)
	if err != nil {
		return nil, err
	}

	waitingValidators, err := nodesCoordinator.NodesInfoToValidators(waitingNodesInfo)
	if err != nil {
		return nil, err
	}

	consensusGroupCache, err := lrucache.NewCache(consensusGroupCacheSize)
	if err != nil {
		return nil, err
	}

	s.memDB = disabled.CreateMemUnit()

	argsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize:    int(args.GenesisNodesConfig.GetShardConsensusGroupSize()),
		MetaConsensusGroupSize:     int(args.GenesisNodesConfig.GetMetaConsensusGroupSize()),
		Marshalizer:                args.Marshalizer,
		Hasher:                     args.Hasher,
		Shuffler:                   args.NodeShuffler,
		EpochStartNotifier:         &disabled.EpochStartNotifier{},
		BootStorer:                 s.memDB,
		ShardIDAsObserver:          args.ShardIdAsObserver,
		NbShards:                   args.GenesisNodesConfig.NumberOfShards(),
		EligibleNodes:              eligibleValidators,
		WaitingNodes:               waitingValidators,
		SelfPublicKey:              args.PubKey,
		ConsensusGroupCache:        consensusGroupCache,
		ShuffledOutHandler:         disabled.NewShuffledOutHandler(),
		WaitingListFixEnabledEpoch: args.WaitingListFixEnableEpoch,
		ChanStopNode:               args.ChanNodeStop,
		NodeTypeProvider:           args.NodeTypeProvider,
		IsFullArchive:              args.IsFullArchive,
	}
	baseNodesCoordinator, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argsNodesCoordinator)
	if err != nil {
		return nil, err
	}

	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, args.ChanceComputer)
	if err != nil {
		return nil, err
	}

	s.nodeCoordinator = nodesCoord

	return s, nil
}

// NodesConfigFromMetaBlock synces and creates registry from epoch start metablock
func (s *syncValidatorStatus) NodesConfigFromMetaBlock(
	currMetaBlock data.HeaderHandler,
	prevMetaBlock data.HeaderHandler,
) (*nodesCoordinator.NodesCoordinatorRegistry, uint32, []*block.MiniBlock, error) {
	if currMetaBlock.GetNonce() > 1 && !currMetaBlock.IsStartOfEpochBlock() {
		return nil, 0, nil, epochStart.ErrNotEpochStartBlock
	}
	if prevMetaBlock.GetNonce() > 1 && !prevMetaBlock.IsStartOfEpochBlock() {
		return nil, 0, nil, epochStart.ErrNotEpochStartBlock
	}

	allMiniblocks := make([]*block.MiniBlock, 0)
	prevMiniBlocks, err := s.processValidatorChangesFor(prevMetaBlock)
	if err != nil {
		return nil, 0, nil, err
	}
	allMiniblocks = append(allMiniblocks, prevMiniBlocks...)

	currentMiniBlocks, err := s.processValidatorChangesFor(currMetaBlock)
	if err != nil {
		return nil, 0, nil, err
	}
	allMiniblocks = append(allMiniblocks, currentMiniBlocks...)

	selfShardId, err := s.nodeCoordinator.ShardIdForEpoch(currMetaBlock.GetEpoch())
	if err != nil {
		return nil, 0, nil, err
	}

	nodesConfig := s.nodeCoordinator.NodesCoordinatorToRegistry()
	nodesConfig.CurrentEpoch = currMetaBlock.GetEpoch()
	return nodesConfig, selfShardId, allMiniblocks, nil
}

func (s *syncValidatorStatus) processValidatorChangesFor(metaBlock data.HeaderHandler) ([]*block.MiniBlock, error) {
	if metaBlock.GetEpoch() == 0 {
		// no need to process for genesis - already created
		return make([]*block.MiniBlock, 0), nil
	}

	blockBody, miniBlocks, err := s.getPeerBlockBodyForMeta(metaBlock)
	if err != nil {
		return nil, err
	}
	s.nodeCoordinator.EpochStartPrepare(metaBlock, blockBody)

	return miniBlocks, nil
}

func findPeerMiniBlockHeaders(metaBlock data.HeaderHandler) []data.MiniBlockHeaderHandler {
	shardMBHeaderHandlers := make([]data.MiniBlockHeaderHandler, 0)
	mbHeaderHandlers := metaBlock.GetMiniBlockHeaderHandlers()
	for i, mbHeader := range mbHeaderHandlers {
		if mbHeader.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		shardMBHeaderHandlers = append(shardMBHeaderHandlers, mbHeaderHandlers[i])
	}
	return shardMBHeaderHandlers
}

func (s *syncValidatorStatus) getPeerBlockBodyForMeta(
	metaBlock data.HeaderHandler,
) (data.BodyHandler, []*block.MiniBlock, error) {
	shardMBHeaders := findPeerMiniBlockHeaders(metaBlock)

	s.miniBlocksSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	err := s.miniBlocksSyncer.SyncPendingMiniBlocks(shardMBHeaders, ctx)
	cancel()
	if err != nil {
		return nil, nil, err
	}

	peerMiniBlocks, err := s.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return nil, nil, err
	}

	blockBody := &block.Body{MiniBlocks: make([]*block.MiniBlock, 0, len(peerMiniBlocks))}
	for _, mbHeader := range shardMBHeaders {
		blockBody.MiniBlocks = append(blockBody.MiniBlocks, peerMiniBlocks[string(mbHeader.GetHash())])
	}

	return blockBody, blockBody.MiniBlocks, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *syncValidatorStatus) IsInterfaceNil() bool {
	return s == nil
}
