package common

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/storage"
)

var log = logger.GetOrCreate("common")

// GetCachedIntermediateTxs will return from cache intermediate transactions
func GetCachedIntermediateTxs(cache storage.Cacher, headerHash []byte) (map[block.Type]map[string]data.TransactionHandler, error) {
	cachedIntermediateTxs, ok := cache.Get(headerHash)
	if !ok {
		log.Warn("intermediateTxs not found in dataPool", "hash", headerHash)
		return nil, fmt.Errorf("%w for header %s", ErrMissingCachedTransactions, hex.EncodeToString(headerHash))
	}

	cachedIntermediateTxsMap, ok := cachedIntermediateTxs.(map[block.Type]map[string]data.TransactionHandler)
	if !ok {
		return nil, fmt.Errorf("%w for cached intermediate transaction %s", ErrWrongTypeAssertion, hex.EncodeToString(headerHash))
	}

	return cachedIntermediateTxsMap, nil
}

// GetCachedLogs will return the cached log events from provided cache
func GetCachedLogs(cache storage.Cacher, headerHash []byte) ([]*data.LogData, error) {
	logsKey := PrepareLogEventsKey(headerHash)
	cachedLogs, ok := cache.Get(logsKey)
	if !ok {
		log.Warn("logs not found in dataPool", "hash", headerHash)
		return nil, fmt.Errorf("%w for header %s", ErrMissingCachedLogs, hex.EncodeToString(headerHash))
	}
	cachedLogsSlice, ok := cachedLogs.([]*data.LogData)
	if !ok {
		return nil, fmt.Errorf("%w for cached logs %s", ErrWrongTypeAssertion, hex.EncodeToString(headerHash))
	}
	return cachedLogsSlice, nil
}

// GetCachedMbs will return the cached miniblocks from provided cache
func GetCachedMbs(cache storage.Cacher, marshaller marshal.Marshalizer, headerHash []byte) ([]*block.MiniBlock, error) {
	cachedIntraMBs, ok := cache.Get(headerHash)
	if !ok {
		log.Warn("intra miniblocks not found in dataPool", "hash", headerHash)
		return nil, fmt.Errorf("%w for header %s", ErrMissingMiniBlock, hex.EncodeToString(headerHash))
	}

	miniBlocks, ok := cachedIntraMBs.([]*block.MiniBlock)
	if !ok {
		return nil, fmt.Errorf("%w for GetCachedMbs", ErrWrongTypeAssertion)
	}

	return miniBlocks, nil
}

// GetCachedBody will return the block body based from provided cache based on the execution result
func GetCachedBody(cache storage.Cacher, marshaller marshal.Marshalizer, baseExecResult data.BaseExecutionResultHandler) (*block.Body, error) {
	miniBlockHeaderHandlers, err := GetMiniBlocksHeaderHandlersFromExecResult(baseExecResult)
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
			return nil, ErrMissingMiniBlock
		}

		cachedMiniBlockBytes := cachedMiniBlock.([]byte)

		var miniBlock *block.MiniBlock
		err = marshaller.Unmarshal(miniBlock, cachedMiniBlockBytes)
		if err != nil {
			return nil, err
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}
