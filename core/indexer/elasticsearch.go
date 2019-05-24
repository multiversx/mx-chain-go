package indexer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin/json"
)

const txBulkSize = 2500
const txIndex = "transactions"

type elasticIndexer struct {
	db *elasticsearch.Client
	marshalizer marshal.Marshalizer
	hasher hashing.Hasher
	logger *logger.Logger
}

// NewElasticIndexer SHOULD UPDATE COMMENT
func NewElasticIndexer(url string, marshalizer marshal.Marshalizer, hasher hashing.Hasher, logger *logger.Logger) (core.Indexer, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	indexer := &elasticIndexer{es, marshalizer, hasher, logger}

	err = indexer.checkAndCreateIndex(txIndex)
	if err != nil {
		return nil, err
	}

	return indexer, nil
}

func (ei *elasticIndexer) checkAndCreateIndex(index string) error {
	res, err := ei.db.Indices.Exists([]string{index})
	if err != nil {
		return err
	}
	// Indices.Exists actually does a HEAD request to the elastic index.
	// A status code of 200 actually means the index exists so we
	//  don't need to do nothing.
	if res.StatusCode == http.StatusOK {
		return nil
	}
	// A status code of 404 means the index does not exist so we create it
	if res.StatusCode == http.StatusNotFound {
		err = ei.createIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) createIndex(index string) error {
	res, err := ei.db.Indices.Create(index)
	if err != nil {
		return err
	}
	if res.IsError() {
		return ErrCannotCreateIndex
	}

	return nil
}

// SaveBlock will build
func (ei *elasticIndexer) SaveBlock(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) {
	// Save Miniblocks
	// Save Block
	// Save Header
	if len(body) == 0 {
		fmt.Println("elasticsearch - no miniblocks")
		return
	}
	ei.saveTransactions(body, header, txPool)
}

func (ei *elasticIndexer) saveTransactions(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) {
	var buff bytes.Buffer

	bulks := ei.buildTransactionBulks(body, header, txPool)

	for _, bulk := range bulks {
		for _, tx := range bulk {
			meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, tx.Hash, "\n"))
			serializedTx, err := json.Marshal(tx)
			if err != nil {
				ei.logger.Warn("could not serialize transaction, will skip indexing: ", tx.Hash)
				continue
			}
			// append a newline foreach element
			serializedTx = append(serializedTx, "\n"...)

			buff.Grow(len(meta) + len(serializedTx))
			buff.Write(meta)
			buff.Write(serializedTx)
		}

		res, err := ei.db.Bulk(bytes.NewReader(buff.Bytes()), ei.db.Bulk.WithIndex(txIndex))
		buff.Reset()

		if err != nil {
			ei.logger.Warn("error indexing bulk of transactions")
		}
		if res.IsError() {
			fmt.Println(res.String())
			ei.logger.Warn("error from elasticsearch indexing bulk of transactions")
		}
	}
}

// buildTransactionBulks creates bulks of maximum txBulkSize transactions to be indexed together
//  using the elasticsearch bulk API
func (ei *elasticIndexer) buildTransactionBulks(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) [][]Transaction {
	processedTxCount := 0
	bulks := make([][]Transaction, (header.GetTxCount() / txBulkSize) + 1)
	blockMarshal, _ := ei.marshalizer.Marshal(body)
	blockHash := ei.hasher.Compute(string(blockMarshal))

	for _, mb := range body {
		mbMarshal, err := ei.marshalizer.Marshal(mb)
		if err != nil {
			ei.logger.Warn("could not marshal miniblock")
			continue
		}
		mbHash := ei.hasher.Compute(string(mbMarshal))

		for _, txHash := range mb.TxHashes {
			processedTxCount++
			currentBulk := processedTxCount / txBulkSize
			currentTx, ok := txPool[string(txHash)]
			if !ok {
				ei.logger.Warn("elasticsearch could not find tx hash in pool")
				continue
			}

			bulks[currentBulk] = append(bulks[currentBulk], Transaction{
				Hash: hex.EncodeToString(txHash),
				MBHash: hex.EncodeToString(mbHash),
				BlockHash: hex.EncodeToString(blockHash),
				Nonce: currentTx.Nonce,
				Value: currentTx.Value,
				Receiver: hex.EncodeToString(currentTx.RcvAddr),
				Sender: hex.EncodeToString(currentTx.SndAddr),
				ReceiverShard: mb.ReceiverShardID,
				SenderShard: mb.SenderShardID,
				GasPrice: currentTx.GasPrice,
				GasLimit: currentTx.GasLimit,
				Data: hex.EncodeToString(currentTx.Data),
				Signature: hex.EncodeToString(currentTx.Signature),
				Timestamp: header.TimeStamp,
			})
		}
	}
	return bulks
}
