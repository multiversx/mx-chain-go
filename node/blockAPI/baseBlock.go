package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type baseAPIBockProcessor struct {
	isFullHistoryNode        bool
	selfShardID              uint32
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	historyProc              fullHistory.HistoryHandler
	unmarshalTx              func(txBytes []byte, txType string) (*transaction.ApiTransactionResult, error)
}

func (bap *baseAPIBockProcessor) getNormalTxsFromMB(miniblock *block.MiniBlock, epoch uint32) []*transaction.ApiTransactionResult {
	txs := make([]*transaction.ApiTransactionResult, len(miniblock.TxHashes))
	unit := dataRetriever.TransactionUnit
	for idx := 0; idx < len(miniblock.TxHashes); idx++ {
		txBytes, err := bap.getFromStorerWithEpoch(unit, miniblock.TxHashes[idx], epoch)
		if err != nil {
			continue
		}

		tx, err := bap.unmarshalTx(txBytes, "normal")
		if err != nil {
			continue
		}

		txs[idx] = tx
	}

	return txs
}

func (bap *baseAPIBockProcessor) getRewardTxsFromMB(miniblock *block.MiniBlock, epoch uint32) []*transaction.ApiTransactionResult {
	txs := make([]*transaction.ApiTransactionResult, len(miniblock.TxHashes))
	unit := dataRetriever.RewardTransactionUnit
	for idx := 0; idx < len(miniblock.TxHashes); idx++ {
		txBytes, err := bap.getFromStorerWithEpoch(unit, miniblock.TxHashes[idx], epoch)
		if err != nil {
			continue
		}

		tx, err := bap.unmarshalTx(txBytes, "rewardTx")
		if err != nil {
			continue
		}

		txs[idx] = tx
	}

	return txs
}

func (bap *baseAPIBockProcessor) getUnsignedTxsFromMB(miniblock *block.MiniBlock, epoch uint32) []*transaction.ApiTransactionResult {
	txs := make([]*transaction.ApiTransactionResult, len(miniblock.TxHashes))
	unit := dataRetriever.UnsignedTransactionUnit
	for idx := 0; idx < len(miniblock.TxHashes); idx++ {
		txBytes, err := bap.getFromStorerWithEpoch(unit, miniblock.TxHashes[idx], epoch)
		if err != nil {
			continue
		}

		tx, err := bap.unmarshalTx(txBytes, "unsignedTx")
		if err != nil {
			continue
		}

		txs[idx] = tx
	}

	return txs
}

func (bap *baseAPIBockProcessor) getTxsByMb(mbHeader *block.MiniBlockHeader, epoch uint32) []*transaction.ApiTransactionResult {
	mbBytes, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, mbHeader.Hash, epoch)
	if err != nil {
		return nil
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, mbBytes)
	if err != nil {
		return nil
	}

	switch miniBlock.Type {
	case block.TxBlock:
		return bap.getNormalTxsFromMB(miniBlock, epoch)
	case block.SmartContractResultBlock:
		return bap.getUnsignedTxsFromMB(miniBlock, epoch)
	case block.RewardsBlock:
		return bap.getRewardTxsFromMB(miniBlock, epoch)
	default:
		return nil
	}
}

func (bap *baseAPIBockProcessor) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
	if !bap.isFullHistoryNode {
		return bap.store.Get(unit, key)
	}

	epoch, err := bap.historyProc.GetEpochForHash(key)
	if err != nil {
		return nil, err
	}

	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBockProcessor) getFromStorerWithEpoch(unit dataRetriever.UnitType, key []byte, epoch uint32) ([]byte, error) {
	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}
