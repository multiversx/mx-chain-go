package block

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	logger "github.com/multiversx/mx-chain-logger-go"
)

func getMetricsFromMetaHeader(
	header *block.MetaBlock,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
	numShardHeadersFromPool int,
	numShardHeadersProcessed uint64,
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

	appStatusHandler.SetUInt64Value(common.MetricHeaderSize, headerSize)
	appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, uint64(header.TxCount))
	appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, numMiniBlocksMetaBlock)
	appStatusHandler.SetUInt64Value(common.MetricNumShardHeadersProcessed, numShardHeadersProcessed)
	appStatusHandler.SetUInt64Value(common.MetricNumShardHeadersFromPool, uint64(numShardHeadersFromPool))
}

func getMetricsFromBlockBody(
	body *block.Body,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
) {
	mbLen := len(body.MiniBlocks)
	miniblocksSize := uint64(0)
	totalTxCount := 0
	for i := 0; i < mbLen; i++ {
		totalTxCount += len(body.MiniBlocks[i].TxHashes)

		marshalizedBlock, err := marshalizer.Marshal(body.MiniBlocks[i])
		if err == nil {
			miniblocksSize += uint64(len(marshalizedBlock))
		}
	}
	appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, uint64(totalTxCount))
	appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, uint64(mbLen))
	appStatusHandler.SetUInt64Value(common.MetricMiniBlocksSize, miniblocksSize)
}

func getMetricsFromHeader(
	header data.HeaderHandler,
	numTxWithDst uint64,
	marshalizer marshal.Marshalizer,
	appStatusHandler core.AppStatusHandler,
) {
	headerSize := uint64(0)
	marshalizedHeader, err := json.Marshal(header)
	log.LogIfError(err)
	log.Debug("marshalized header", "header bytes", string(marshalizedHeader))

	marshalizedHeader, err = marshalizer.Marshal(header)
	if err == nil {
		headerSize = uint64(len(marshalizedHeader))
	}

	appStatusHandler.SetUInt64Value(common.MetricHeaderSize, headerSize)
	appStatusHandler.SetUInt64Value(common.MetricTxPoolLoad, numTxWithDst)
}

func saveMetricsForCommittedShardBlock(
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	appStatusHandler core.AppStatusHandler,
	currentBlockHash string,
	highestFinalBlockNonce uint64,
	metaBlock data.HeaderHandler,
	shardHeader data.HeaderHandler,
) {
	incrementCountAcceptedBlocks(nodesCoordinator, appStatusHandler, shardHeader)
	appStatusHandler.SetUInt64Value(common.MetricEpochNumber, uint64(shardHeader.GetEpoch()))
	appStatusHandler.SetStringValue(common.MetricCurrentBlockHash, currentBlockHash)
	appStatusHandler.SetUInt64Value(common.MetricHighestFinalBlock, highestFinalBlockNonce)
	appStatusHandler.SetStringValue(common.MetricCrossCheckBlockHeight, fmt.Sprintf("meta %d", metaBlock.GetNonce()))
	appStatusHandler.SetUInt64Value(common.MetricCrossCheckBlockHeightMeta, metaBlock.GetNonce())
}

func incrementCountAcceptedBlocks(
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	appStatusHandler core.AppStatusHandler,
	header data.HeaderHandler,
) {
	consensusGroup, err := nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch(),
	)
	if err != nil {
		return
	}

	ownPubKey := nodesCoordinator.GetOwnPublicKey()
	myIndex := 0
	found := false
	for idx, val := range consensusGroup {
		if bytes.Equal(ownPubKey, val.PubKey()) {
			myIndex = idx
			found = true
			break
		}
	}

	if !found {
		return
	}

	bitMap := header.GetPubKeysBitmap()
	indexOutOfBounds := myIndex/8 >= len(bitMap)
	if indexOutOfBounds {
		log.Trace("process blocks metrics: index out of bounds",
			"index", myIndex,
			"bitMap", bitMap,
			"bitMap length", len(bitMap))
		return
	}

	indexInBitmap := myIndex != 0 && bitMap[myIndex/8]&(1<<uint8(myIndex%8)) != 0
	if indexInBitmap {
		appStatusHandler.Increment(common.MetricCountConsensusAcceptedBlocks)
	}
}

func saveMetricsForCommitMetachainBlock(
	appStatusHandler core.AppStatusHandler,
	header *block.MetaBlock,
	headerHash []byte,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	highestFinalBlockNonce uint64,
) {
	appStatusHandler.SetStringValue(common.MetricCurrentBlockHash, logger.DisplayByteSlice(headerHash))
	appStatusHandler.SetUInt64Value(common.MetricEpochNumber, uint64(header.Epoch))
	appStatusHandler.SetUInt64Value(common.MetricHighestFinalBlock, highestFinalBlockNonce)

	// TODO: remove if epoch start block needs to be validated by the new epoch nodes
	epoch := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epoch > 0 {
		epoch = epoch - 1
	}

	pubKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(
		header.PrevRandSeed,
		header.Round,
		core.MetachainShardId,
		epoch,
	)
	if err != nil {
		log.Debug("cannot get validators public keys", "error", err.Error())
	}

	countMetaAcceptedSignedBlocks(pubKeys, nodesCoordinator.GetOwnPublicKey(), appStatusHandler)
}

func countMetaAcceptedSignedBlocks(
	publicKeys []string,
	ownPublicKey []byte,
	appStatusHandler core.AppStatusHandler,
) {
	isInConsensus := false

	for index, publicKey := range publicKeys {
		if bytes.Equal([]byte(publicKey), ownPublicKey) {
			if index == 0 {
				return
			}

			isInConsensus = true
			break
		}
	}

	if !isInConsensus {
		return
	}

	appStatusHandler.Increment(common.MetricCountConsensusAcceptedBlocks)
}

func indexRoundInfo(
	outportHandler outport.OutportHandler,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	shardId uint32,
	header data.HeaderHandler,
	lastHeader data.HeaderHandler,
	signersIndexes []uint64,
) {
	roundInfo := &outportcore.RoundInfo{
		Round:            header.GetRound(),
		SignersIndexes:   signersIndexes,
		BlockWasProposed: true,
		ShardId:          shardId,
		Epoch:            header.GetEpoch(),
		Timestamp:        uint64(time.Duration(header.GetTimeStamp())),
	}

	if check.IfNil(lastHeader) {
		outportHandler.SaveRoundsInfo(&outportcore.RoundsInfo{RoundsInfo: []*outportcore.RoundInfo{roundInfo}})
		return
	}

	lastBlockRound := lastHeader.GetRound()
	currentBlockRound := header.GetRound()
	roundDuration := calculateRoundDuration(lastHeader.GetTimeStamp(), header.GetTimeStamp(), lastBlockRound, currentBlockRound)

	roundsInfo := make([]*outportcore.RoundInfo, 0)
	roundsInfo = append(roundsInfo, roundInfo)
	for i := lastBlockRound + 1; i < currentBlockRound; i++ {
		publicKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(lastHeader.GetRandSeed(), i, shardId, lastHeader.GetEpoch())
		if err != nil {
			continue
		}
		signersIndexes, err = nodesCoordinator.GetValidatorsIndexes(publicKeys, lastHeader.GetEpoch())
		if err != nil {
			log.Error(err.Error(), "round", i)
			continue
		}

		roundInfo = &outportcore.RoundInfo{
			Round:            i,
			SignersIndexes:   signersIndexes,
			BlockWasProposed: false,
			ShardId:          shardId,
			Epoch:            header.GetEpoch(),
			Timestamp:        uint64(time.Duration(header.GetTimeStamp() - ((currentBlockRound - i) * roundDuration))),
		}

		roundsInfo = append(roundsInfo, roundInfo)
	}

	outportHandler.SaveRoundsInfo(&outportcore.RoundsInfo{RoundsInfo: roundsInfo})
}

func indexValidatorsRating(
	outportHandler outport.OutportHandler,
	valStatProc process.ValidatorStatisticsProcessor,
	metaBlock data.HeaderHandler,
) {
	// TODO use validatorInfoProvider  to get information about rating
	latestHash, err := valStatProc.RootHash()
	if err != nil {
		return
	}

	validators, err := valStatProc.GetValidatorInfoForRootHash(latestHash)
	if err != nil {
		return
	}

	for shardID, validatorInfosInShard := range validators {
		validatorsInfos := make([]*outportcore.ValidatorRatingInfo, 0)
		for _, validatorInfo := range validatorInfosInShard {
			validatorsInfos = append(validatorsInfos, &outportcore.ValidatorRatingInfo{
				PublicKey: hex.EncodeToString(validatorInfo.PublicKey),
				Rating:    float32(validatorInfo.Rating) * 100 / 10000000,
			})
		}

		outportHandler.SaveValidatorsRating(&outportcore.ValidatorsRating{
			ShardID:              shardID,
			Epoch:                metaBlock.GetEpoch(),
			ValidatorsRatingInfo: validatorsInfos,
		})
	}
}

func calculateRoundDuration(
	lastBlockTimestamp uint64,
	currentBlockTimestamp uint64,
	lastBlockRound uint64,
	currentBlockRound uint64,
) uint64 {
	if lastBlockTimestamp >= currentBlockTimestamp {
		log.Debug("last block timestamp is greater or equals than current block timestamp")
		return 0
	}
	if lastBlockRound >= currentBlockRound {
		log.Debug("last block round is greater or equals than current block round")
		return 0
	}

	diffTimeStamp := currentBlockTimestamp - lastBlockTimestamp
	diffRounds := currentBlockRound - lastBlockRound

	return diffTimeStamp / diffRounds
}

func saveEpochStartEconomicsMetrics(statusHandler core.AppStatusHandler, epochStartMetaBlock *block.MetaBlock) {
	economics := epochStartMetaBlock.EpochStart.Economics

	statusHandler.SetStringValue(common.MetricTotalSupply, economics.TotalSupply.String())
	statusHandler.SetStringValue(common.MetricInflation, economics.TotalNewlyMinted.String())
	statusHandler.SetStringValue(common.MetricTotalFees, epochStartMetaBlock.AccumulatedFeesInEpoch.String())
	statusHandler.SetStringValue(common.MetricDevRewardsInEpoch, epochStartMetaBlock.DevFeesInEpoch.String())
	statusHandler.SetUInt64Value(common.MetricEpochForEconomicsData, uint64(epochStartMetaBlock.Epoch))
}
