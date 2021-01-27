package process

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

func prepareGeneralInfo(tpsBenchmark statistics.TPSBenchmark) bytes.Buffer {
	var buff bytes.Buffer

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, metachainTpsDocID, tpsIndex, "\n"))
	generalInfo := types.TPS{
		LiveTPS:               tpsBenchmark.LiveTPS(),
		PeakTPS:               tpsBenchmark.PeakTPS(),
		NrOfShards:            tpsBenchmark.NrOfShards(),
		BlockNumber:           tpsBenchmark.BlockNumber(),
		RoundNumber:           tpsBenchmark.RoundNumber(),
		RoundTime:             tpsBenchmark.RoundTime(),
		AverageBlockTxCount:   tpsBenchmark.AverageBlockTxCount(),
		LastBlockTxCount:      tpsBenchmark.LastBlockTxCount(),
		TotalProcessedTxCount: tpsBenchmark.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(generalInfo)
	if err != nil {
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this round")
		return buff
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		log.Warn("elastic search: update TPS write meta", "error", err.Error())
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		log.Warn("elastic search: update TPS write serialized info", "error", err.Error())
	}

	return buff
}

func serializeShardInfo(shardInfo statistics.ShardStatistic) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s%d", "_type" : "%s" } }%s`,
		shardTpsDocIDPrefix, shardInfo.ShardID(), tpsIndex, "\n"))

	bigTxCount := big.NewInt(int64(shardInfo.AverageBlockTxCount()))
	shardTPS := types.TPS{
		ShardID:               shardInfo.ShardID(),
		LiveTPS:               shardInfo.LiveTPS(),
		PeakTPS:               shardInfo.PeakTPS(),
		AverageTPS:            shardInfo.AverageTPS(),
		AverageBlockTxCount:   bigTxCount,
		CurrentBlockNonce:     shardInfo.CurrentBlockNonce(),
		LastBlockTxCount:      shardInfo.LastBlockTxCount(),
		TotalProcessedTxCount: shardInfo.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(shardTPS)
	if err != nil {
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this shard")
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

func serializeValidatorsRating(index string, validatorsRatingInfo []types.ValidatorRatingInfo) ([]bytes.Buffer, error) {
	var err error
	var buff bytes.Buffer

	buffSlice := make([]bytes.Buffer, 0)
	for _, valRatingInfo := range validatorsRatingInfo {
		id := fmt.Sprintf("%s_%s", valRatingInfo.PublicKey, index)
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))
		serializedData, errPrepareSc := json.Marshal(valRatingInfo)
		if errPrepareSc != nil {
			log.Warn("indexer: marshal",
				"error", "could not serialize validator rating information, will skip indexing",
				"hash", valRatingInfo.PublicKey)
			continue
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentValidatorRating := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentValidatorRating > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk of validators rating information, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk of validators rating information", "error", err.Error())
			return nil, err
		}

	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

func serializeRoundInfo(info types.RoundInfo) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d_%d", "_type" : "%s" } }%s`,
		info.ShardId, info.Index, "_doc", "\n"))

	serializedInfo, err := json.Marshal(info)
	if err != nil {
		log.Warn("indexer: could not serialize round info, will skip indexing this round info",
			"error", err.Error())
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

func serializeEpochInfoData(header data.HeaderHandler) ([]byte, error) {
	metablock, ok := header.(*block.MetaBlock)
	if !ok {
		return nil, ErrHeaderTypeAssertion
	}

	epochInfo := &types.EpochInfo{
		AccumulatedFees: metablock.AccumulatedFeesInEpoch.String(),
		DeveloperFees:   metablock.DevFeesInEpoch.String(),
	}

	epochInfoBytes, err := json.Marshal(epochInfo)
	if err != nil {
		return nil, err
	}

	return epochInfoBytes, nil
}
