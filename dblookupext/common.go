package dblookupext

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

func getIntermediateTxs(cache storage.Cacher, headerHash []byte) (map[string]data.TransactionHandler, map[string]data.TransactionHandler, error) {
	cachedIntermediateTxs, ok := cache.Get(headerHash)
	if !ok {
		log.Warn("intermediateTxs not found in dataPool", "hash", headerHash)
		return nil, nil, fmt.Errorf("%w for header %s", process.ErrMissingHeader, hex.EncodeToString(headerHash))
	}

	cachedIntermediateTxsMap, ok := cachedIntermediateTxs.(map[block.Type]map[string]data.TransactionHandler)
	if !ok {
		return make(map[string]data.TransactionHandler), make(map[string]data.TransactionHandler), nil
	}

	scrs := cachedIntermediateTxsMap[block.SmartContractResultBlock]
	receipts := cachedIntermediateTxsMap[block.ReceiptBlock]

	return scrs, receipts, nil
}

func getLogs(cache storage.Cacher, headerHash []byte) ([]*data.LogData, error) {
	logsKey := common.PrepareLogEventsKey(headerHash)
	cachedLogs, ok := cache.Get(logsKey)
	if !ok {
		log.Warn("logs not found in dataPool", "hash", headerHash)
		return nil, fmt.Errorf("%w for header %s", process.ErrMissingHeader, hex.EncodeToString(headerHash))
	}
	cachedLogsSlice, ok := cachedLogs.([]*data.LogData)
	if !ok {
		return []*data.LogData{}, nil
	}
	return cachedLogsSlice, nil
}

func getIntraMbs(cache storage.Cacher, marshaller marshal.Marshalizer, headerHash []byte) ([]*block.MiniBlock, error) {
	cachedIntraMBs, ok := cache.Get(headerHash)
	if !ok {
		log.Warn("intra miniblocks not found in dataPool", "hash", headerHash)
		return nil, fmt.Errorf("%w for header %s", process.ErrMissingHeader, hex.EncodeToString(headerHash))
	}
	cachedLogsBuff := cachedIntraMBs.([]byte)
	var intraMBs []*block.MiniBlock
	errUnmarshal := marshaller.Unmarshal(&intraMBs, cachedLogsBuff)
	if errUnmarshal != nil {
		return nil, fmt.Errorf("%w getIntraMbs: cannot unmarshall", errUnmarshal)
	}

	return intraMBs, nil
}

func getBody(cache storage.Cacher, marshaller marshal.Marshalizer, baseExecResult data.BaseExecutionResultHandler, shardID uint32) (*block.Body, error) {
	miniBlockHeaderHandlers, err := extractMiniBlocksHeaderHandlersFromExecResult(baseExecResult, shardID)
	if err != nil {
		return nil, err
	}

	var miniBlocks block.MiniBlockSlice
	for _, miniBlockHeaderHandler := range miniBlockHeaderHandlers {
		mbHash := miniBlockHeaderHandler.GetHash()
		cachedMiniBlock, found := cache.Get(mbHash)
		if !found {
			log.Warn("mini block from execution result not cached after execution",
				"mini block hash", mbHash)
			return nil, process.ErrMissingMiniBlock
		}

		cachedMiniBlockBytes := cachedMiniBlock.([]byte)

		var miniBlock *block.MiniBlock
		err = marshaller.Unmarshal(&miniBlock, cachedMiniBlockBytes)
		if err != nil {
			return nil, err
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

func extractMiniBlocksHeaderHandlersFromExecResult(
	baseExecResult data.BaseExecutionResultHandler,
	headerShard uint32,
) ([]data.MiniBlockHeaderHandler, error) {
	if headerShard == common.MetachainShardId {
		metaExecResult, ok := baseExecResult.(data.MetaExecutionResultHandler)
		if !ok {
			log.Warn("extractMiniBlocksHeaderHandlersFromExecResult assert failed to MetaExecutionResultHandler")
			return nil, process.ErrWrongTypeAssertion
		}

		return metaExecResult.GetMiniBlockHeadersHandlers(), nil
	}

	execResult, ok := baseExecResult.(data.ExecutionResultHandler)
	if !ok {
		log.Warn("extractMiniBlocksHeaderHandlersFromExecResult assert failed to ExecutionResultHandler")
		return nil, process.ErrWrongTypeAssertion
	}

	return execResult.GetMiniBlockHeadersHandlers(), nil
}
