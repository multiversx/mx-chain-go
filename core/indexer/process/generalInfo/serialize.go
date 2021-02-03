package generalInfo

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

// SerializeGeneralInfo will serialize general information
func (gip *generalInfoProcessor) SerializeGeneralInfo(genInfo *types.TPS, shardsInfo []*types.TPS, index string) *bytes.Buffer {
	buff := serializeGeneralInfo(genInfo, index)

	for _, shardInfo := range shardsInfo {
		serializeShardInfo(buff, shardInfo, index)
	}

	return buff
}

func serializeGeneralInfo(generalInfo *types.TPS, index string) *bytes.Buffer {
	buff := &bytes.Buffer{}
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, metachainTpsDocID, index, "\n"))

	serializedInfo, err := json.Marshal(generalInfo)
	if err != nil {
		log.Debug("serializeGeneralInfo could not serialize tps info", "error", err)
		return buff
	}

	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		log.Warn("serializeGeneralInfo cannot write meta", "error", err)
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		log.Warn("serializeGeneralInfo cannot write serialized info", "error", err)
	}

	return buff
}

func serializeShardInfo(buff *bytes.Buffer, shardTPS *types.TPS, index string) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s%d", "_type" : "%s" } }%s`,
		shardTpsDocIDPrefix, shardTPS.ShardID, index, "\n"))

	serializedInfo, err := json.Marshal(shardTPS)
	if err != nil {
		log.Debug("serializeShardInfo could not serialize info", "error", err)
		return
	}

	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		log.Warn("serializeShardInfo cannot write meta", "error", err)
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		log.Warn("serializeShardInfo cannot write serialized info", "error", err)
	}
}

// SerializeRoundsInfo will serialize information about rounds
func (gip *generalInfoProcessor) SerializeRoundsInfo(roundsInfo []types.RoundInfo) *bytes.Buffer {
	buff := &bytes.Buffer{}
	for _, info := range roundsInfo {
		serializedRoundInfo, meta := serializeRoundInfo(info)

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("generalInfoProcessor.SaveRoundsInfo cannot write meta", "error", err)
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("generalInfoProcessor.SaveRoundsInfo cannot write serialized round info", "error", err)
		}
	}

	return buff
}

func serializeRoundInfo(info types.RoundInfo) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d_%d", "_type" : "%s" } }%s`,
		info.ShardId, info.Index, "_doc", "\n"))

	serializedInfo, err := json.Marshal(info)
	if err != nil {
		log.Warn("serializeRoundInfo could not serialize round info, will skip indexing this round info", "error", err)
		return nil, nil
	}

	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}
