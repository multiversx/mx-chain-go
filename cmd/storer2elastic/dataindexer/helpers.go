package dataindexer

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	dataIndexerDisabled "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/dataindexer/disabled"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

func (dpi *dataProcessor) computeNotarizedHeaders(hdr data.HeaderHandler) []string {
	metaBlock, ok := hdr.(*block.MetaBlock)
	if !ok {
		return []string{}
	}

	numShardInfo := len(metaBlock.ShardInfo)
	notarizedHdrs := make([]string, 0, numShardInfo)
	for _, shardInfo := range metaBlock.ShardInfo {
		notarizedHdrs = append(notarizedHdrs, hex.EncodeToString(shardInfo.HeaderHash))
	}

	return notarizedHdrs
}

func (dpi *dataProcessor) processValidatorsForEpoch(epoch uint32, metaBlock *block.MetaBlock, mbUnit storage.Persister) {
	peerMiniBlocks := make([]*block.MiniBlock, 0)

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.Type != block.PeerBlock {
			continue
		}

		mbBytes, err := mbUnit.Get(mbHeader.Hash)
		if err != nil {
			log.Warn("cannot find peer mini block in storage", "hash", mbHeader.Hash)
			continue
		}
		recoveredMiniBlock := &block.MiniBlock{}
		err = dpi.marshalizer.Unmarshal(recoveredMiniBlock, mbBytes)
		if err != nil {
			log.Warn("cannot unmarshal peer miniblock", "error", err)
			continue
		}

		peerMiniBlocks = append(peerMiniBlocks, recoveredMiniBlock)
	}

	peerBlock := &block.Body{
		MiniBlocks: peerMiniBlocks,
	}

	for shardID := range dpi.nodesCoordinators {
		dpi.nodesCoordinators[shardID].EpochStartPrepare(metaBlock, peerBlock)
	}
}

func (dpi *dataProcessor) canIndexHeaderNow(hdr data.HeaderHandler) bool {
	shardID := hdr.GetShardID()
	nodesCoord, ok := dpi.nodesCoordinators[shardID]
	if !ok {
		return false
	}

	testPubKeys := make([]string, 0)
	_, err := nodesCoord.GetValidatorsIndexes(testPubKeys, hdr.GetEpoch())
	if err == nil {
		return true
	}

	return false
}

func (dpi *dataProcessor) computeSignersIndexes(hdr data.HeaderHandler) ([]uint64, error) {
	nodesCoordinator, ok := dpi.nodesCoordinators[hdr.GetShardID()]
	if !ok {
		return nil, fmt.Errorf("nodes coordinator not found for shard %d", hdr.GetShardID())
	}

	publicKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(
		hdr.GetPrevRandSeed(), hdr.GetRound(), hdr.GetShardID(), hdr.GetEpoch(),
	)
	if err != nil {
		return nil, err
	}

	return nodesCoordinator.GetValidatorsIndexes(publicKeys, hdr.GetEpoch())
}

func (dpi *dataProcessor) createNodesCoordinators(nodesConfig sharding.GenesisNodesSetupHandler) (map[uint32]NodesCoordinator, error) {
	nodesCoordinatorsMap := make(map[uint32]NodesCoordinator)
	shardIDs := dpi.getShardIDs()
	for _, shardID := range shardIDs {
		nodeCoordForShard, err := dpi.createNodesCoordinatorForShard(nodesConfig, shardID)
		if err != nil {
			return nil, err
		}
		nodesCoordinatorsMap[shardID] = nodeCoordForShard
	}

	return nodesCoordinatorsMap, nil
}

func (dpi *dataProcessor) createNodesCoordinatorForShard(nodesConfig sharding.GenesisNodesSetupHandler, shardID uint32) (NodesCoordinator, error) {
	eligibleNodesInfo, waitingNodesInfo := nodesConfig.InitialNodesInfo()

	eligibleValidators, err := sharding.NodesInfoToValidators(eligibleNodesInfo)
	if err != nil {
		return nil, err
	}

	waitingValidators, err := sharding.NodesInfoToValidators(waitingNodesInfo)
	if err != nil {
		return nil, err
	}

	consensusGroupCache, err := lrucache.NewCache(int(nodesConfig.GetShardConsensusGroupSize()))
	if err != nil {
		return nil, err
	}

	memDB := disabled.CreateMemUnit()

	argsNodesCoordinator := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: int(nodesConfig.GetShardConsensusGroupSize()),
		MetaConsensusGroupSize:  int(nodesConfig.GetMetaConsensusGroupSize()),
		Marshalizer:             dpi.marshalizer,
		Hasher:                  dpi.hasher,
		Shuffler:                dataIndexerDisabled.NewNodesShuffler(),
		EpochStartNotifier:      &disabled.EpochStartNotifier{},
		BootStorer:              memDB,
		ShardIDAsObserver:       shardID,
		NbShards:                nodesConfig.NumberOfShards(),
		EligibleNodes:           eligibleValidators,
		WaitingNodes:            waitingValidators,
		SelfPublicKey:           []byte("own public key"),
		ConsensusGroupCache:     consensusGroupCache,
		ShuffledOutHandler:      disabled.NewShuffledOutHandler(),
	}
	baseNodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argsNodesCoordinator)
	if err != nil {
		return nil, fmt.Errorf("%w while creating nodes coordinator", err)
	}

	return baseNodesCoordinator, nil
}

func (dpi *dataProcessor) preparePersistersHolder(dbInfo *databasereader.DatabaseInfo) (*persistersHolder, error) {
	persHold := &persistersHolder{}

	miniBlocksPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "MiniBlocks")
	if err != nil {
		return nil, err
	}
	persHold.miniBlocksPersister = miniBlocksPersister

	txsPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "Transactions")
	if err != nil {
		return nil, err
	}
	persHold.transactionPersister = txsPersister

	uTxsPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "UnsignedTransactions")
	if err != nil {
		return nil, err
	}
	persHold.unsignedTransactionsPersister = uTxsPersister

	rTxsPersister, err := dpi.databaseReader.LoadPersister(dbInfo, "RewardTransactions")
	if err != nil {
		return nil, err
	}
	persHold.rewardTransactionsPersister = rTxsPersister

	return persHold, nil
}

func (dpi *dataProcessor) closePersisters(persisters *persistersHolder) {
	err := persisters.miniBlocksPersister.Close()
	log.LogIfError(err)

	err = persisters.transactionPersister.Close()
	log.LogIfError(err)

	err = persisters.unsignedTransactionsPersister.Close()
	log.LogIfError(err)

	err = persisters.rewardTransactionsPersister.Close()
	log.LogIfError(err)
}
