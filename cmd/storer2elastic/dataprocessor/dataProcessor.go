package dataprocessor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	storer2ElasticData "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/data"
	dataProcessorDisabled "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/dataprocessor/disabled"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

var log = logger.GetOrCreate("dataprocessor")

// ArgsDataProcessor holds the arguments needed for creating a new dataProcessor
type ArgsDataProcessor struct {
	ElasticIndexer      indexer.Indexer
	DataReplayer        DataReplayerHandler
	GenesisNodesSetup   sharding.GenesisNodesSetupHandler
	ShardCoordinator    sharding.Coordinator
	Marshalizer         marshal.Marshalizer
	Hasher              hashing.Hasher
	TPSBenchmarkUpdater TPSBenchmarkUpdaterHandler
	RatingsProcessor    *ratingsProcessor
}

type dataProcessor struct {
	elasticIndexer      indexer.Indexer
	dataReplayer        DataReplayerHandler
	genesisNodesSetup   sharding.GenesisNodesSetupHandler
	shardCoordinator    sharding.Coordinator
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	nodesCoordinators   map[uint32]NodesCoordinator
	tpsBenchmarkUpdater TPSBenchmarkUpdaterHandler
	ratingsProcessor    *ratingsProcessor
}

// NewDataProcessor returns a new instance of dataProcessor
func NewDataProcessor(args ArgsDataProcessor) (*dataProcessor, error) {
	if check.IfNil(args.ElasticIndexer) {
		return nil, ErrNilElasticIndexer
	}
	if check.IfNil(args.DataReplayer) {
		return nil, ErrNilDataReplayer
	}
	if check.IfNil(args.GenesisNodesSetup) {
		return nil, ErrNilGenesisNodesSetup
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.TPSBenchmarkUpdater) {
		return nil, ErrNilTPSBenchmarkUpdater
	}

	dp := &dataProcessor{
		elasticIndexer:      args.ElasticIndexer,
		dataReplayer:        args.DataReplayer,
		genesisNodesSetup:   args.GenesisNodesSetup,
		shardCoordinator:    args.ShardCoordinator,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		ratingsProcessor:    args.RatingsProcessor,
		tpsBenchmarkUpdater: args.TPSBenchmarkUpdater,
	}

	nodesCoordinators, err := dp.createNodesCoordinators(args.GenesisNodesSetup)
	if err != nil {
		return nil, err
	}

	dp.nodesCoordinators = nodesCoordinators

	return dp, nil
}

// Index will range over data from storage and will index it
func (dp *dataProcessor) Index() error {
	return dp.dataReplayer.Range(dp.processData)
}

func (dp *dataProcessor) processData(persistedData storer2ElasticData.RoundPersistedData) bool {
	metaPersistedData := persistedData.MetaBlockData
	if metaPersistedData.Header.IsStartOfEpochBlock() || metaPersistedData.Header.GetNonce() == 0 {
		metaBlock, _ := metaPersistedData.Header.(*block.MetaBlock)
		dp.processValidatorsForEpoch(metaBlock, metaPersistedData.Body)
		err := dp.ratingsProcessor.IndexRatingsForEpochStartMetaBlock(metaBlock)
		if err != nil {
			log.Error("cannot process ratings", "error", err)
		}
	}

	err := dp.indexData(metaPersistedData)
	if err != nil {
		log.Warn("error indexing header", "error", err)
	}
	metaBlock, _ := metaPersistedData.Header.(*block.MetaBlock)
	dp.tpsBenchmarkUpdater.IndexTPSForMetaBlock(metaBlock)

	for _, shardData := range persistedData.ShardHeaders {
		err = dp.indexData(shardData)
		if err != nil {
			log.Warn("error indexing shard header", "error", err)
		}
	}

	return true
}

func (dp *dataProcessor) indexData(data *storer2ElasticData.HeaderData) error {
	signersIndexes, err := dp.computeSignersIndexes(data.Header)
	if err != nil {
		return err
	}

	notarizedHeaders := dp.computeNotarizedHeaders(data.Header)
	newBody := &block.Body{MiniBlocks: make([]*block.MiniBlock, 0)}
	for _, mb := range data.Body.MiniBlocks {
		if mb.Type == block.ReceiptBlock { // don't index receipt miniblocks
			continue
		}

		newBody.MiniBlocks = append(newBody.MiniBlocks, mb)
	}
	dp.elasticIndexer.SaveBlock(newBody, data.Header, data.BodyTransactions, signersIndexes, notarizedHeaders)
	dp.indexRoundInfo(signersIndexes, data.Header)
	dp.logHeaderInfo(data.Header)
	return nil
}

func (dp *dataProcessor) indexRoundInfo(signersIndexes []uint64, hdr data.HeaderHandler) {
	ri := indexer.RoundInfo{
		Index:            hdr.GetRound(),
		SignersIndexes:   signersIndexes,
		BlockWasProposed: false,
		ShardId:          hdr.GetShardID(),
		Timestamp:        time.Duration(hdr.GetTimeStamp()),
	}

	dp.elasticIndexer.SaveRoundsInfos([]indexer.RoundInfo{ri})
}

func (dp *dataProcessor) computeSignersIndexes(hdr data.HeaderHandler) ([]uint64, error) {
	nodesCoordinator, ok := dp.nodesCoordinators[hdr.GetShardID()]
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

func (dp *dataProcessor) createNodesCoordinators(nodesConfig sharding.GenesisNodesSetupHandler) (map[uint32]NodesCoordinator, error) {
	nodesCoordinatorsMap := make(map[uint32]NodesCoordinator)
	shardIDs := dp.getShardIDs()
	for _, shardID := range shardIDs {
		nodeCoordForShard, err := dp.createNodesCoordinatorForShard(nodesConfig, shardID)
		if err != nil {
			return nil, err
		}
		nodesCoordinatorsMap[shardID] = nodeCoordForShard
	}

	return nodesCoordinatorsMap, nil
}

func (dp *dataProcessor) createNodesCoordinatorForShard(nodesConfig sharding.GenesisNodesSetupHandler, shardID uint32) (NodesCoordinator, error) {
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
		Marshalizer:             dp.marshalizer,
		Hasher:                  dp.hasher,
		Shuffler:                dataProcessorDisabled.NewNodesShuffler(),
		EpochStartNotifier:      disabled.NewEpochStartNotifier(),
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

func (dp *dataProcessor) getShardIDs() []uint32 {
	shardIDs := make([]uint32, 0)
	for shard := uint32(0); shard < dp.shardCoordinator.NumberOfShards(); shard++ {
		shardIDs = append(shardIDs, shard)
	}
	shardIDs = append(shardIDs, core.MetachainShardId)

	return shardIDs
}

func (dp *dataProcessor) processValidatorsForEpoch(metaBlock *block.MetaBlock, body *block.Body) {
	if metaBlock.Epoch == 0 {
		return
	}

	peerMiniBlocks := make([]*block.MiniBlock, 0)

	for _, mb := range body.MiniBlocks {
		if mb.Type != block.PeerBlock {
			continue
		}

		mbHash, err := core.CalculateHash(dp.marshalizer, dp.hasher, mb)
		if err != nil {
			continue
		}

		for _, hash := range metaBlock.MiniBlockHeaders {
			if bytes.Equal(hash.Hash, mbHash) {
				peerMiniBlocks = append(peerMiniBlocks, mb)
				break
			}
		}

	}

	peerMiniBlocks = dp.uniqueMiniBlocksSlice(peerMiniBlocks)

	log.Warn("length of peer block", "len", len(peerMiniBlocks))
	peerBlock := &block.Body{
		MiniBlocks: peerMiniBlocks,
	}

	for shardID := range dp.nodesCoordinators {
		dp.nodesCoordinators[shardID].EpochStartPrepare(metaBlock, peerBlock)
	}

	validatorsPubKeys, err := dp.nodesCoordinators[core.MetachainShardId].GetAllEligibleValidatorsPublicKeys(metaBlock.Epoch)
	if err != nil || len(validatorsPubKeys) == 0 {
		log.Warn("cannot get all eligible validatorsPubKeys", "epoch", metaBlock.Epoch)
		return
	}

	dp.elasticIndexer.SaveValidatorsPubKeys(validatorsPubKeys, metaBlock.Epoch)
}

func (dp *dataProcessor) uniqueMiniBlocksSlice(mbs []*block.MiniBlock) []*block.MiniBlock {
	keys := make(map[string]bool)
	list := make([]*block.MiniBlock, 0)
	for _, entry := range mbs {
		hash, err := core.CalculateHash(dp.marshalizer, dp.hasher, entry)
		if err != nil {
			continue
		}

		_, valueOk := keys[string(hash)]
		if !valueOk {
			keys[string(hash)] = true
			list = append(list, entry)
		}
	}
	return list
}

func (dp *dataProcessor) computeNotarizedHeaders(hdr data.HeaderHandler) []string {
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

func (dp *dataProcessor) logHeaderInfo(hdr data.HeaderHandler) {
	headerHash, err := core.CalculateHash(dp.marshalizer, dp.hasher, hdr)
	if err != nil {
		log.Warn("error while calculating the hash of a header for logging", "error", err)
		return
	}

	log.Info("indexed header",
		"epoch", hdr.GetEpoch(),
		"shard", hdr.GetShardID(),
		"nonce", hdr.GetNonce(),
		"hash", headerHash,
	)
}
