package workItems

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("core/indexer/workItems")

type itemBlock struct {
	indexer                saveBlockIndexer
	marshalizer            marshal.Marshalizer
	bodyHandler            data.BodyHandler
	headerHandler          data.HeaderHandler
	txPool                 map[string]data.TransactionHandler
	signersIndexes         []uint64
	notarizedHeadersHashes []string
	headerHash             []byte
}

// NewItemBlock will create a new instance of ItemBlock
func NewItemBlock(
	indexer saveBlockIndexer,
	marshalizer marshal.Marshalizer,
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
	notarizedHeadersHashes []string,
	headerHash []byte,
) WorkItemHandler {
	return &itemBlock{
		indexer:                indexer,
		bodyHandler:            bodyHandler,
		headerHandler:          headerHandler,
		txPool:                 txPool,
		signersIndexes:         signersIndexes,
		notarizedHeadersHashes: notarizedHeadersHashes,
		marshalizer:            marshalizer,
		headerHash:             headerHash,
	}
}

// Save will prepare and save a block item in elasticsearch database
func (wib *itemBlock) Save() error {
	if check.IfNil(wib.headerHandler) {
		log.Warn("nil header provided when trying to index block, will skip")
		return nil
	}

	log.Debug("indexer: starting indexing block", "hash", wib.headerHash, "nonce", wib.headerHandler.GetNonce())

	body, ok := wib.bodyHandler.(*block.Body)
	if !ok {
		return fmt.Errorf("%w when trying body assertion, block hash %s, nonce %d",
			ErrBodyTypeAssertion, logger.DisplayByteSlice(wib.headerHash), wib.headerHandler.GetNonce())
	}

	txsSizeInBytes := ComputeSizeOfTxs(wib.marshalizer, wib.txPool)
	err := wib.indexer.SaveHeader(wib.headerHandler, wib.signersIndexes, body, wib.notarizedHeadersHashes, txsSizeInBytes)
	if err != nil {
		return fmt.Errorf("%w when saving header block, hash %s, nonce %d",
			err, logger.DisplayByteSlice(wib.headerHash), wib.headerHandler.GetNonce())
	}

	if len(body.MiniBlocks) == 0 {
		return nil
	}

	mbsInDb, err := wib.indexer.SaveMiniblocks(wib.headerHandler, body)
	if err != nil {
		return fmt.Errorf("%w when saving miniblocks, block hash %s, nonce %d",
			err, logger.DisplayByteSlice(wib.headerHash), wib.headerHandler.GetNonce())
	}

	shardID := wib.headerHandler.GetShardID()
	err = wib.indexer.SaveTransactions(body, wib.headerHandler, wib.txPool, shardID, mbsInDb)
	if err != nil {
		return fmt.Errorf("%w when saving transactions, block hash %s, nonce %d",
			err, logger.DisplayByteSlice(wib.headerHash), wib.headerHandler.GetNonce())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wib *itemBlock) IsInterfaceNil() bool {
	return wib == nil
}

// ComputeSizeOfTxs will compute size of transactions in bytes
func ComputeSizeOfTxs(marshalizer marshal.Marshalizer, txs map[string]data.TransactionHandler) int {
	if len(txs) == 0 {
		return 0
	}

	txsSize := 0
	for _, tx := range txs {
		txBytes, err := marshalizer.Marshal(tx)
		if err != nil {
			log.Debug("indexer: marshal transaction", "error", err)
			continue
		}

		txsSize += len(txBytes)
	}

	return txsSize
}
