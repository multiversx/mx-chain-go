package bootstrap

import (
	"bytes"
	"fmt"

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
	nodeCoordinator  sharding.NodesCoordinator
}

// ArgsNewSyncValidatorStatus
type ArgsNewSyncValidatorStatus struct {
	DataPool       dataRetriever.PoolsHolder
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
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

	nodesConfig := &sharding.NodesCoordinatorRegistry{
		EpochsConfig: make(map[string]*sharding.EpochValidators),
		CurrentEpoch: currMetaBlock.Epoch,
	}

	epochValidators, selfShardId, err := s.processNodesConfigFor(currMetaBlock, publicKey)
	if err != nil {
		return nil, 0, err
	}
	configId := fmt.Sprint(currMetaBlock.Epoch)
	nodesConfig.EpochsConfig[configId] = epochValidators

	epochValidators, _, err = s.processNodesConfigFor(prevMetaBlock, nil)
	if err != nil {
		return nil, 0, err
	}
	configId = fmt.Sprint(prevMetaBlock.Epoch)
	nodesConfig.EpochsConfig[configId] = epochValidators

	// TODO: use shuffling process from nodesCoordinator to create the final data

	return nodesConfig, selfShardId, nil
}

func (s *syncValidatorStatus) processNodesConfigFor(
	metaBlock *block.MetaBlock,
	publicKey []byte,
) (*sharding.EpochValidators, uint32, error) {
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
		return nil, 0, err
	}

	peerMiniBlocks, err := s.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return nil, 0, err
	}

	epochValidators := &sharding.EpochValidators{
		EligibleValidators: make(map[string][]*sharding.SerializableValidator),
		WaitingValidators:  make(map[string][]*sharding.SerializableValidator),
	}

	selfShardId := core.AllShardId
	found := false
	shouldSearchSelfId := len(publicKey) == 0
	for _, mb := range peerMiniBlocks {
		for _, txHash := range mb.TxHashes {
			vid := &state.ValidatorInfo{}
			err := s.marshalizer.Unmarshal(vid, txHash)
			if err != nil {
				return nil, 0, err
			}

			serializableValidator := &sharding.SerializableValidator{
				PubKey:  vid.PublicKey,
				Address: vid.RewardAddress, // TODO - take out - need to refactor validator.go and its usage across the project
			}

			shardId := fmt.Sprint(vid.ShardId)
			switch vid.List {
			case string(core.EligibleList):
				epochValidators.EligibleValidators[shardId] = append(epochValidators.EligibleValidators[shardId], serializableValidator)
			case string(core.WaitingList):
				epochValidators.WaitingValidators[shardId] = append(epochValidators.EligibleValidators[shardId], serializableValidator)
			}

			if shouldSearchSelfId && !found && bytes.Equal(vid.PublicKey, publicKey) {
				selfShardId = vid.ShardId
				found = true
			}
		}
	}

	return epochValidators, selfShardId, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *syncValidatorStatus) IsInterfaceNil() bool {
	return s == nil
}
