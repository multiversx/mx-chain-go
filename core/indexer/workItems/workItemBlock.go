package workItems

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
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
	txIndexingEnabled      bool
}

// NewItemBlock will create a new instance of ItemBlock
func NewItemBlock(
	indexer saveBlockIndexer,
	marshalizer marshal.Marshalizer,
	txIndexingEnabled bool,
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
	notarizedHeadersHashes []string,
) WorkItemHandler {
	return &itemBlock{
		txIndexingEnabled:      txIndexingEnabled,
		indexer:                indexer,
		bodyHandler:            bodyHandler,
		headerHandler:          headerHandler,
		txPool:                 txPool,
		signersIndexes:         signersIndexes,
		notarizedHeadersHashes: notarizedHeadersHashes,
		marshalizer:            marshalizer,
	}
}

// Save will prepare and save a block item in elasticsearch database
func (wib *itemBlock) Save() error {
	body, ok := wib.bodyHandler.(*block.Body)
	if !ok {
		log.Warn("itemBlock.Save", "body", ErrBodyTypeAssertion.Error())
		return ErrBodyTypeAssertion
	}

	txsSizeInBytes := computeSizeOfTxs(wib.marshalizer, wib.txPool)
	err := wib.indexer.SaveHeader(wib.headerHandler, wib.signersIndexes, body, wib.notarizedHeadersHashes, txsSizeInBytes)
	if err != nil {
		log.Warn("itemBlock.Save", "could not index block:", err.Error())
		return err
	}

	if len(body.MiniBlocks) == 0 {
		return nil
	}

	mbsInDb, err := wib.indexer.SaveMiniblocks(wib.headerHandler, body)
	if err != nil {
		log.Warn("itemBlock.Save", "could not index miniblocks", err.Error())
		return err
	}

	if !wib.txIndexingEnabled {
		return nil
	}

	shardID := wib.headerHandler.GetShardID()
	err = wib.indexer.SaveTransactions(body, wib.headerHandler, wib.txPool, shardID, mbsInDb)
	if err != nil {
		log.Warn("itemBlock.Save", "could not index transactions", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wib *itemBlock) IsInterfaceNil() bool {
	return wib == nil
}

func computeSizeOfTxs(marshalizer marshal.Marshalizer, txs map[string]data.TransactionHandler) int {
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
