package bootstrap

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
	DataPool           dataRetriever.PoolsHolder
	Marshalizer        marshal.Marshalizer
	Hasher             hashing.Hasher
	RequestHandler     process.RequestHandler
	ChanceComputer     sharding.ChanceComputer
	GenesisNodesConfig sharding.GenesisNodesSetupHandler
	NodeShuffler       sharding.NodesShuffler
	PubKey             []byte
	ShardIdAsObserver  uint32
}

// NewSyncValidatorStatus creates a new validator status process component
func NewSyncValidatorStatus(args ArgsNewSyncValidatorStatus) (*syncValidatorStatus, error) {
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

	eligibleValidators, err := sharding.NodesInfoToValidators(eligibleNodesInfo)
	if err != nil {
		return nil, err
	}

	waitingValidators, err := sharding.NodesInfoToValidators(waitingNodesInfo)
	if err != nil {
		return nil, err
	}

	consensusGroupCache, err := lrucache.NewCache(consensusGroupCacheSize)
	if err != nil {
		return nil, err
	}

	s.memDB = disabled.CreateMemUnit()

	argsNodesCoordinator := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: int(args.GenesisNodesConfig.GetShardConsensusGroupSize()),
		MetaConsensusGroupSize:  int(args.GenesisNodesConfig.GetMetaConsensusGroupSize()),
		Marshalizer:             args.Marshalizer,
		Hasher:                  args.Hasher,
		Shuffler:                args.NodeShuffler,
		EpochStartNotifier:      &disabled.EpochStartNotifier{},
		BootStorer:              s.memDB,
		ShardIDAsObserver:       args.ShardIdAsObserver,
		NbShards:                args.GenesisNodesConfig.NumberOfShards(),
		EligibleNodes:           eligibleValidators,
		WaitingNodes:            waitingValidators,
		SelfPublicKey:           args.PubKey,
		ConsensusGroupCache:     consensusGroupCache,
		ShuffledOutHandler:      disabled.NewShuffledOutHandler(),
		EpochNotifier:           disabled.NewEpochNotifier(),
	}
	baseNodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argsNodesCoordinator)
	if err != nil {
		return nil, err
	}

	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, args.ChanceComputer)
	if err != nil {
		return nil, err
	}

	s.nodeCoordinator = nodesCoordinator

	return s, nil
}

// NodesConfigFromMetaBlock synces and creates registry from epoch start metablock
func (s *syncValidatorStatus) NodesConfigFromMetaBlock(
	currMetaBlock *block.MetaBlock,
	prevMetaBlock *block.MetaBlock,
) (*sharding.NodesCoordinatorRegistry, uint32, error) {
	if currMetaBlock.GetNonce() > 1 && !currMetaBlock.IsStartOfEpochBlock() {
		return nil, 0, epochStart.ErrNotEpochStartBlock
	}
	if prevMetaBlock.GetNonce() > 1 && !prevMetaBlock.IsStartOfEpochBlock() {
		return nil, 0, epochStart.ErrNotEpochStartBlock
	}

	err := s.processValidatorChangesFor(prevMetaBlock)
	if err != nil {
		return nil, 0, err
	}

	err = s.processValidatorChangesFor(currMetaBlock)
	if err != nil {
		return nil, 0, err
	}

	selfShardId, err := s.nodeCoordinator.ShardIdForEpoch(currMetaBlock.Epoch)
	if err != nil {
		return nil, 0, err
	}

	nodesConfig := s.nodeCoordinator.NodesCoordinatorToRegistry()
	nodesConfig.CurrentEpoch = currMetaBlock.Epoch
	return nodesConfig, selfShardId, nil
}

func (s *syncValidatorStatus) processValidatorChangesFor(metaBlock *block.MetaBlock) error {
	if metaBlock.GetEpoch() == 0 {
		// no need to process for genesis - already created
		return nil
	}

	blockBody, err := s.getPeerBlockBodyForMeta(metaBlock)
	if err != nil {
		return err
	}
	s.nodeCoordinator.EpochStartPrepare(metaBlock, blockBody)

	return nil
}

func findPeerMiniBlockHeaders(metaBlock *block.MetaBlock) []block.MiniBlockHeader {
	shardMBHeaders := make([]block.MiniBlockHeader, 0)
	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.Type != block.PeerBlock {
			continue
		}

		shardMBHdr := block.MiniBlockHeader{
			Hash:            mbHeader.Hash,
			ReceiverShardID: mbHeader.ReceiverShardID,
			SenderShardID:   core.MetachainShardId,
			TxCount:         mbHeader.TxCount,
			Type:            mbHeader.Type,
		}
		shardMBHeaders = append(shardMBHeaders, shardMBHdr)
	}
	return shardMBHeaders
}

func (s *syncValidatorStatus) getPeerBlockBodyForMeta(
	metaBlock *block.MetaBlock,
) (data.BodyHandler, error) {
	shardMBHeaders := findPeerMiniBlockHeaders(metaBlock)

	s.miniBlocksSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	err := s.miniBlocksSyncer.SyncPendingMiniBlocks(shardMBHeaders, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	peerMiniBlocks, err := s.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return nil, err
	}

	blockBody := &block.Body{MiniBlocks: make([]*block.MiniBlock, 0, len(peerMiniBlocks))}
	for _, mbHeader := range shardMBHeaders {
		blockBody.MiniBlocks = append(blockBody.MiniBlocks, peerMiniBlocks[string(mbHeader.Hash)])
	}

	return blockBody, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *syncValidatorStatus) IsInterfaceNil() bool {
	return s == nil
}
