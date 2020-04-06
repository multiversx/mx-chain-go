package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update/sync"
)

type syncValidatorStatus struct {
	miniBlocksSyncer   epochStart.PendingMiniBlocksSyncHandler
	dataPool           dataRetriever.PoolsHolder
	marshalizer        marshal.Marshalizer
	requestHandler     process.RequestHandler
	nodeCoordinator    EpochStartNodesCoordinator
	genesisNodesConfig *sharding.NodesSetup
}

// ArgsNewSyncValidatorStatus
type ArgsNewSyncValidatorStatus struct {
	DataPool           dataRetriever.PoolsHolder
	Marshalizer        marshal.Marshalizer
	RequestHandler     process.RequestHandler
	Rater              sharding.ChanceComputer
	GenesisNodesConfig *sharding.NodesSetup
	NodeShuffler       sharding.NodesShuffler
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

	argsNodesCoordinator := ArgsNewStartInEpochNodesCoordinator{
		Shuffler:                args.NodeShuffler,
		Chance:                  args.Rater,
		ShardConsensusGroupSize: args.GenesisNodesConfig.ConsensusGroupSize,
		MetaConsensusGroupSize:  args.GenesisNodesConfig.MetaChainConsensusGroupSize,
	}
	s.nodeCoordinator, err = NewStartInEpochNodesCoordinator(argsNodesCoordinator)

	return s, nil
}

// NodesConfigFromMetaBlock synces and creates registry from epoch start metablock
func (s *syncValidatorStatus) NodesConfigFromMetaBlock(
	currMetaBlock *block.MetaBlock,
	prevMetaBlock *block.MetaBlock,
	publicKey []byte,
) (*sharding.NodesCoordinatorRegistry, uint32, error) {
	if !currMetaBlock.IsStartOfEpochBlock() {
		return nil, 0, epochStart.ErrNotEpochStartBlock
	}
	if !prevMetaBlock.IsStartOfEpochBlock() {
		return nil, 0, epochStart.ErrNotEpochStartBlock
	}

	prevEpochsValidators, err := s.computeNodesConfigFor(prevMetaBlock)
	if err != nil {
		return nil, 0, err
	}

	currEpochsValidators, err := s.computeNodesConfigFor(currMetaBlock)
	if err != nil {
		return nil, 0, err
	}

	selfShardId := s.nodeCoordinator.ComputeShardForSelfPublicKey(currMetaBlock.Epoch, publicKey)

	nodesConfig := &sharding.NodesCoordinatorRegistry{
		EpochsConfig: make(map[string]*sharding.EpochValidators, 2),
		CurrentEpoch: currMetaBlock.Epoch,
	}

	epochConfigId := fmt.Sprint(prevMetaBlock.Epoch)
	nodesConfig.EpochsConfig[epochConfigId] = prevEpochsValidators
	epochConfigId = fmt.Sprint(currMetaBlock.Epoch)
	nodesConfig.EpochsConfig[epochConfigId] = currEpochsValidators

	return nodesConfig, selfShardId, nil
}

func (s *syncValidatorStatus) computeNodesConfigFor(metaBlock *block.MetaBlock) (*sharding.EpochValidators, error) {
	if metaBlock.Epoch == 0 {
		return s.nodeCoordinator.ComputeNodesConfigForGenesis(s.genesisNodesConfig)
	}

	epochValidatorsInfo, err := s.processNodesConfigFor(metaBlock)
	if err != nil {
		return nil, err
	}

	return s.nodeCoordinator.ComputeNodesConfigFor(metaBlock, epochValidatorsInfo)
}

func findPeerMiniBlockHeaders(metaBlock *block.MetaBlock) []block.ShardMiniBlockHeader {
	shardMBHeaders := make([]block.ShardMiniBlockHeader, 0)
	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.Type != block.PeerBlock {
			continue
		}

		shardMBHdr := block.ShardMiniBlockHeader{
			Hash:            mbHeader.Hash,
			ReceiverShardID: mbHeader.ReceiverShardID,
			SenderShardID:   core.MetachainShardId,
			TxCount:         mbHeader.TxCount,
		}
		shardMBHeaders = append(shardMBHeaders, shardMBHdr)
	}
	return shardMBHeaders
}

func (s *syncValidatorStatus) processNodesConfigFor(
	metaBlock *block.MetaBlock,
) ([]*state.ShardValidatorInfo, error) {
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

	validatorInfos := make([]*state.ShardValidatorInfo, 0)
	for _, mb := range peerMiniBlocks {
		for _, txHash := range mb.TxHashes {
			vid := &state.ShardValidatorInfo{}
			err := s.marshalizer.Unmarshal(vid, txHash)
			if err != nil {
				return nil, err
			}

			validatorInfos = append(validatorInfos, vid)
		}
	}

	return validatorInfos, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *syncValidatorStatus) IsInterfaceNil() bool {
	return s == nil
}
