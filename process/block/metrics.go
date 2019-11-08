package block

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler/persister"
)

func getMetricsFromMetaHeader(
	header *block.MetaBlock,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
	headersCountInPool int,
	totalHeadersProcessed uint64,
) {
	numMiniBlocksMetaBlock := uint64(0)
	headerSize := uint64(0)

	for _, shardInfo := range header.ShardInfo {
		numMiniBlocksMetaBlock += uint64(len(shardInfo.ShardMiniBlockHeaders))
	}

	marshalizedHeader, err := marshalizer.Marshal(header)
	if err == nil {
		headerSize = uint64(len(marshalizedHeader))
	}

	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, headerSize)
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(header.TxCount))
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, numMiniBlocksMetaBlock)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersProcessed, totalHeadersProcessed)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersFromPool, uint64(headersCountInPool))
}

func getMetricsFromBlockBody(
	body block.Body,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
) {
	mbLen := len(body)
	miniblocksSize := uint64(0)
	totalTxCount := 0
	for i := 0; i < mbLen; i++ {
		totalTxCount += len(body[i].TxHashes)

		marshalizedBlock, err := marshalizer.Marshal(body[i])
		if err == nil {
			miniblocksSize += uint64(len(marshalizedBlock))
		}
	}
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(totalTxCount))
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, uint64(mbLen))
	appStatusHandler.SetUInt64Value(core.MetricMiniBlocksSize, miniblocksSize)
}

func getMetricsFromHeader(
	header *block.Header,
	numTxWithDst uint64,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
) {
	headerSize := uint64(0)
	marshalizedHeader, err := marshalizer.Marshal(header)
	if err == nil {
		headerSize = uint64(len(marshalizedHeader))
	}

	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, headerSize)
	appStatusHandler.SetUInt64Value(core.MetricTxPoolLoad, numTxWithDst)
}

func saveMetricsForACommittedBlock(
	appStatusHandler core.AppStatusHandler,
	isInConsensus bool,
	currentBlockHash string,
	highestFinalBlockNonce uint64,
	headerMetaNonce uint64,
) {
	if isInConsensus {
		appStatusHandler.Increment(core.MetricCountConsensusAcceptedBlocks)
	}
	appStatusHandler.SetStringValue(core.MetricCurrentBlockHash, currentBlockHash)
	appStatusHandler.SetUInt64Value(core.MetricHighestFinalBlockInShard, highestFinalBlockNonce)
	appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, fmt.Sprintf("meta %d", headerMetaNonce))
}

func saveMetachainCommitBlockMetrics(
	appStatusHandler core.AppStatusHandler,
	header *block.MetaBlock,
	headerHash []byte,
	nodesCoordinator sharding.NodesCoordinator,

) {
	appStatusHandler.SetStringValue(core.MetricCurrentBlockHash, core.ToB64(headerHash))

	pubKeys, err := nodesCoordinator.GetValidatorsPublicKeys(header.PrevRandSeed, header.Round, sharding.MetachainShardId)
	if err != nil {
		log.Error("cannot get validators public keys", err)
	}

	countNotarizedHeaders(pubKeys, nodesCoordinator.GetOwnPublicKey(), appStatusHandler, len(header.ShardInfo))
}

func countNotarizedHeaders(
	publicKeys []string,
	ownPublicKey []byte,
	appStatusHandler core.AppStatusHandler,
	numBlockHeaders int,
) {
	isInConsensus := false

	for _, publicKey := range publicKeys {
		if bytes.Equal([]byte(publicKey), ownPublicKey) {
			isInConsensus = true
			break
		}
	}

	if !isInConsensus || numBlockHeaders == 0 {
		return
	}

	appStatusHandler.AddUint64(core.MetricCountConsensusAcceptedBlocks, uint64(numBlockHeaders))
}

func saveRoundInfoInElastic(
	elasticIndexer indexer.Indexer,
	nodesCoordinator sharding.NodesCoordinator,
	shardId uint32,
	header data.HeaderHandler,
	lastHeader data.HeaderHandler,
	signersIndexes []uint64,
) {
	roundInfo := indexer.RoundInfo{
		Index:            header.GetRound(),
		SignersIndexes:   signersIndexes,
		BlockWasProposed: true,
		ShardId:          shardId,
		Timestamp:        time.Duration(header.GetTimeStamp()),
	}

	go elasticIndexer.SaveRoundInfo(roundInfo)

	if lastHeader == nil {
		return
	}

	lastBlockRound := lastHeader.GetRound()
	currentBlockRound := header.GetRound()
	roundDuration := calculateRoundDuration(lastHeader.GetTimeStamp(), header.GetTimeStamp(), lastBlockRound, currentBlockRound)
	for i := lastBlockRound + 1; i < currentBlockRound; i++ {
		publicKeys, err := nodesCoordinator.GetValidatorsPublicKeys(lastHeader.GetRandSeed(), i, shardId)
		if err != nil {
			continue
		}
		signersIndexes = nodesCoordinator.GetValidatorsIndexes(publicKeys)
		roundInfo = indexer.RoundInfo{
			Index:            i,
			SignersIndexes:   signersIndexes,
			BlockWasProposed: false,
			ShardId:          shardId,
			Timestamp:        time.Duration(header.GetTimeStamp() - ((currentBlockRound - i) * roundDuration)),
		}

		go elasticIndexer.SaveRoundInfo(roundInfo)
	}
}

func calculateRoundDuration(
	lastBlockTimestamp uint64,
	currentBlockTimestamp uint64,
	lastBlockRound uint64,
	currentBlockRound uint64,
) uint64 {
	if lastBlockTimestamp >= currentBlockTimestamp {
		log.Error("last block timestamp is greater or equals than current block timestamp")
		return 0
	}
	if lastBlockRound >= currentBlockRound {
		log.Error("last block round is greater or equals than current block round")
		return 0
	}

	diffTimeStamp := currentBlockTimestamp - lastBlockTimestamp
	diffRounds := currentBlockRound - lastBlockRound

	return diffTimeStamp / diffRounds
}

func getNumObjFromStorage(store dataRetriever.StorageService, marshalizer marshal.Marshalizer, metricName string) uint64 {
	if store == nil || marshalizer == nil {
		return 0
	}

	dataBytes, err := store.Get(dataRetriever.StatusMetricsUnit, []byte(persister.StatusMetricsDbEntry))
	if err != nil {
		return 0
	}

	var dataFromDb map[string]interface{}
	err = marshalizer.Unmarshal(&dataFromDb, dataBytes)
	if err != nil {
		return 0
	}

	numObj, ok := dataFromDb[metricName].(float64)
	if !ok {
		return 0
	}

	return uint64(numObj)
}
