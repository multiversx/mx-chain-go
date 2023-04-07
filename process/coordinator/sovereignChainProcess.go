package coordinator

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

type sovereignChainTransactionCoordinator struct {
	*transactionCoordinator
}

// NewSovereignChainTransactionCoordinator creates a new sovereign chain transaction coordinator
func NewSovereignChainTransactionCoordinator(
	trasnsactionCoordinator *transactionCoordinator,
) (*sovereignChainTransactionCoordinator, error) {
	if trasnsactionCoordinator == nil {
		return nil, process.ErrNilTransactionCoordinator
	}

	sctc := &sovereignChainTransactionCoordinator{
		transactionCoordinator: trasnsactionCoordinator,
	}

	return sctc, nil
}

// CreateMbsAndProcessCrossShardTransactionsDstMe creates mini blocks and processes cross shard transactions with destination in self shard
func (sctc *sovereignChainTransactionCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
) (block.MiniBlockSlice, uint32, bool, error) {
	shardHeaderExtendedHanlder, isShardHeaderExtendedHandler := hdr.(data.ShardHeaderExtendedHandler)
	if !isShardHeaderExtendedHandler {
		log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: wrong type assertion from data.HeaderHandler to data.ShardHeaderExtendedHandler")
		return make(block.MiniBlockSlice, 0), 0, false, nil
	}

	log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe",
		"schedule mode", scheduledMode,
		"header round", shardHeaderExtendedHanlder.GetRound(),
		"header nonce", shardHeaderExtendedHanlder.GetNonce(),
		"num incoming mbs", len(shardHeaderExtendedHanlder.GetIncomingMiniBlockHandlers()))

	for _, mbh := range shardHeaderExtendedHanlder.GetIncomingMiniBlockHandlers() {
		mb, isMiniBlock := mbh.(*block.MiniBlock)
		if !isMiniBlock {
			log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: wrong type assertion from data.MiniBlockHandler to *block.MiniBlock")
			continue
		}

		log.Debug("incoming mini block",
			"mb type", mb.Type,
			"sender shard", mb.SenderShardID,
			"receiver shard", mb.ReceiverShardID,
			"num incoming txs", len(mb.TxHashes))

		for _, txHash := range mb.TxHashes {
			log.Debug("incoming tx", "hash", txHash)
		}
	}

	//TODO: Implement functionality for sovereign chain
	return make(block.MiniBlockSlice, 0), 0, false, nil
}
