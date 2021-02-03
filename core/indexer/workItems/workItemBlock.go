package workItems

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("core/indexer/workItems")

type itemBlock struct {
	indexer       saveBlockIndexer
	marshalizer   marshal.Marshalizer
	argsSaveBlock *types.ArgsSaveBlockData
}

// NewItemBlock will create a new instance of ItemBlock
func NewItemBlock(
	indexer saveBlockIndexer,
	marshalizer marshal.Marshalizer,
	args *types.ArgsSaveBlockData,
) WorkItemHandler {
	return &itemBlock{
		indexer:       indexer,
		marshalizer:   marshalizer,
		argsSaveBlock: args,
	}
}

// Save will prepare and save a block item in elasticsearch database
func (wib *itemBlock) Save() error {
	if check.IfNil(wib.argsSaveBlock.Header) {
		log.Warn("nil header provided when trying to index block, will skip")
		return nil
	}

	log.Debug("indexer: starting indexing block",
		"hash", logger.DisplayByteSlice(wib.argsSaveBlock.HeaderHash),
		"nonce", wib.argsSaveBlock.Header.GetNonce())

	body, ok := wib.argsSaveBlock.Body.(*block.Body)
	if !ok {
		return fmt.Errorf("%w when trying body assertion, block hash %s, nonce %d",
			ErrBodyTypeAssertion, logger.DisplayByteSlice(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	txsSizeInBytes := ComputeSizeOfTxs(wib.marshalizer, wib.argsSaveBlock.TransactionsPool)
	err := wib.indexer.SaveHeader(wib.argsSaveBlock.Header, wib.argsSaveBlock.SignersIndexes, body, wib.argsSaveBlock.NotarizedHeadersHashes, txsSizeInBytes)
	if err != nil {
		return fmt.Errorf("%w when saving header block, hash %s, nonce %d",
			err, logger.DisplayByteSlice(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	if len(body.MiniBlocks) == 0 {
		return nil
	}

	mbsInDb, err := wib.indexer.SaveMiniblocks(wib.argsSaveBlock.Header, body)
	if err != nil {
		return fmt.Errorf("%w when saving miniblocks, block hash %s, nonce %d",
			err, logger.DisplayByteSlice(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	err = wib.indexer.SaveTransactions(body, wib.argsSaveBlock.Header, wib.argsSaveBlock.TransactionsPool, mbsInDb)
	if err != nil {
		return fmt.Errorf("%w when saving transactions, block hash %s, nonce %d",
			err, logger.DisplayByteSlice(wib.argsSaveBlock.HeaderHash), wib.argsSaveBlock.Header.GetNonce())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wib *itemBlock) IsInterfaceNil() bool {
	return wib == nil
}

// ComputeSizeOfTxs will compute size of transactions in bytes
func ComputeSizeOfTxs(marshalizer marshal.Marshalizer, pool *types.Pool) int {
	sizeTxs := 0

	sizeTxs += computeSizeOfMap(marshalizer, pool.Txs)
	sizeTxs += computeSizeOfMap(marshalizer, pool.Receipts)
	sizeTxs += computeSizeOfMap(marshalizer, pool.Invalid)
	sizeTxs += computeSizeOfMap(marshalizer, pool.Rewards)
	sizeTxs += computeSizeOfMap(marshalizer, pool.Scrs)

	return sizeTxs
}

func computeSizeOfMap(marshalizer marshal.Marshalizer, mapTxs map[string]data.TransactionHandler) int {
	txsSize := 0
	for _, tx := range mapTxs {
		txBytes, err := marshalizer.Marshal(tx)
		if err != nil {
			log.Debug("itemBlock.computeSizeOfMap", "error", err)
			continue
		}

		txsSize += len(txBytes)
	}

	return txsSize
}
