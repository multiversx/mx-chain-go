package bootstrap

import (
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
	miniBlocksSyncer epochStart.PendingMiniBlocksSyncHandler
	dataPool         dataRetriever.PoolsHolder
	marshalizer      marshal.Marshalizer
	requestHandler   process.RequestHandler
	nodeCoordinator  EpochStartNodesCoordinator
}

// ArgsNewSyncValidatorStatus
type ArgsNewSyncValidatorStatus struct {
	DataPool            dataRetriever.PoolsHolder
	Marshalizer         marshal.Marshalizer
	RequestHandler      process.RequestHandler
	Rater               sharding.ChanceComputer
	GenesisNodesConfig  *sharding.NodesSetup
	ValidatorAccountsDB state.AccountsAdapter
	AdrConv             state.AddressConverter
}

// NewSyncValidatorStatus creates a new validator status process component
func NewSyncValidatorStatus(args ArgsNewSyncValidatorStatus) (*syncValidatorStatus, error) {
	s := &syncValidatorStatus{
		dataPool:       args.DataPool,
		marshalizer:    args.Marshalizer,
		requestHandler: args.RequestHandler,
	}
	syncMiniBlocksArgs := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        &disabled.Storer{},
		Cache:          s.dataPool.MiniBlocks(),
		Marshalizer:    s.marshalizer,
		RequestHandler: s.requestHandler,
	}
	var err error
	s.miniBlocksSyncer, err = sync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
	if err != nil {
		return nil, err
	}

	nodeShuffler := sharding.NewXorValidatorsShuffler(
		args.GenesisNodesConfig.MinNodesPerShard,
		args.GenesisNodesConfig.MetaChainMinNodes,
		args.GenesisNodesConfig.Hysteresis,
		args.GenesisNodesConfig.Adaptivity,
	)

	argsNodesCoordinator := ArgsNewStartInEpochNodesCoordinator{
		Shuffler:                nodeShuffler,
		Chance:                  args.Rater,
		ShardConsensusGroupSize: args.GenesisNodesConfig.ConsensusGroupSize,
		MetaConsensusGroupSize:  args.GenesisNodesConfig.MetaChainConsensusGroupSize,
		AdrConv:                 args.AdrConv,
		ValidatorAccountsDB:     args.ValidatorAccountsDB,
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

	validatorInfos := make(map[uint32][]*state.ValidatorInfo)

	epochValidators, err := s.processNodesConfigFor(currMetaBlock)
	if err != nil {
		return nil, 0, err
	}
	validatorInfos[currMetaBlock.Epoch] = epochValidators

	prevEpochValidators, err := s.processNodesConfigFor(prevMetaBlock)
	if err != nil {
		return nil, 0, err
	}
	validatorInfos[prevMetaBlock.Epoch] = prevEpochValidators

	return nodesConfig, selfShardId, nil
}

func (s *syncValidatorStatus) processNodesConfigFor(
	metaBlock *block.MetaBlock,
) ([]*state.ValidatorInfo, error) {
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

	s.miniBlocksSyncer.ClearFields()
	err := s.miniBlocksSyncer.SyncPendingMiniBlocks(shardMBHeaders, timeToWait)
	if err != nil {
		return nil, err
	}

	peerMiniBlocks, err := s.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return nil, err
	}

	validatorInfos := make([]*state.ValidatorInfo, 0)
	for _, mb := range peerMiniBlocks {
		for _, txHash := range mb.TxHashes {
			vid := &state.ValidatorInfo{}
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
