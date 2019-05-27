package indexer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gin-gonic/gin/json"
)

const txBulkSize = 2500
const txIndex = "transactions"
const blockIndex = "blocks"

type elasticIndexer struct {
	db               *elasticsearch.Client
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	logger           *logger.Logger
}

// NewElasticIndexer SHOULD UPDATE COMMENT
func NewElasticIndexer(url string, shardCoordinator sharding.Coordinator, marshalizer marshal.Marshalizer,
	hasher hashing.Hasher, logger *logger.Logger) (core.Indexer, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	indexer := &elasticIndexer{es, shardCoordinator,
		marshalizer, hasher, logger}

	err = indexer.checkAndCreateIndex(blockIndex)
	if err != nil {
		return nil, err
	}

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
		fmt.Println("Status code", res.String())
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
		ei.logger.Warn(res.String())
		return ErrCannotCreateIndex
	}

	return nil
}

// SaveBlock will build
func (ei *elasticIndexer) SaveBlock(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) {
	go ei.saveHeader(header)

	if len(body) == 0 {
		fmt.Println("elasticsearch - no miniblocks")
		return
	}
	go ei.saveTransactions(body, header, txPool)
}

func (ei *elasticIndexer) saveHeader(header *block.Header) {
	var buff bytes.Buffer

	h, err := ei.marshalizer.Marshal(header)
	if err != nil {
		ei.logger.Warn("could not marshal header")
		return
	}

	headerHash := ei.hasher.Compute(string(h))
	elasticBlock := Block{
		Nonce: header.Nonce,
		ShardID: header.ShardId,
		Hash: hex.EncodeToString(headerHash),
		// TODO: We should add functionality for proposer and validators
		Proposer: hex.EncodeToString([]byte("mock proposer")),
		//Validators: "mock validators",
		PubKeyBitmap: hex.EncodeToString(header.PubKeysBitmap),
		Size: int64(len(h)),
		Timestamp: time.Duration(header.TimeStamp),
		TxCount: header.TxCount,
		StateRootHash: hex.EncodeToString(header.RootHash),
		PrevHash: hex.EncodeToString(header.PrevHash),
	}
	serializedBlock, err := json.Marshal(elasticBlock)
	if err != nil {
		ei.logger.Warn("could not marshal elastic header")
		return
	}
	buff.Grow(len(serializedBlock))
	buff.Write(serializedBlock)

	req := esapi.IndexRequest{
		Index: blockIndex,
		DocumentID: hex.EncodeToString(headerHash),
		Body: bytes.NewReader(buff.Bytes()),
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), ei.db)
	if err != nil {
		ei.logger.Warn("Could not index block header: %s", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.IsError() {
		fmt.Println(res.String())
		ei.logger.Warn("error from elasticsearch indexing bulk of transactions")
	}
}

func (ei *elasticIndexer) saveTransactions(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) {
	var buff bytes.Buffer

	bulks := ei.buildTransactionBulks(body, header, txPool)

	for _, bulk := range bulks {
		for _, tx := range bulk {
			meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, tx.Hash, txIndex, "\n"))
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
	bulks := make([][]Transaction, (header.GetTxCount()/txBulkSize)+1)
	blockMarshal, _ := ei.marshalizer.Marshal(body)
	blockHash := ei.hasher.Compute(string(blockMarshal))

	for _, mb := range body {
		mbMarshal, err := ei.marshalizer.Marshal(mb)
		if err != nil {
			ei.logger.Warn("could not marshal miniblock")
			continue
		}
		mbHash := ei.hasher.Compute(string(mbMarshal))

		mbTxStatus := "Pending"
		if ei.shardCoordinator.SelfId() == mb.ReceiverShardID {
			mbTxStatus = "Success"
		}

		for _, txHash := range mb.TxHashes {
			processedTxCount++

			currentBulk := processedTxCount / txBulkSize
			currentTx, ok := txPool[string(txHash)]
			if !ok {
				ei.logger.Warn("elasticsearch could not find tx hash in pool")
				continue
			}

			if ei.shardCoordinator.SelfId() == mb.SenderShardID {

			}
			bulks[currentBulk] = append(bulks[currentBulk], Transaction{
				Hash:              hex.EncodeToString(txHash),
				MBHash:            hex.EncodeToString(mbHash),
				BlockHash:         hex.EncodeToString(blockHash),
				Nonce:             currentTx.Nonce,
				Value:             currentTx.Value,
				Receiver:          hex.EncodeToString(currentTx.RcvAddr),
				Sender:            hex.EncodeToString(currentTx.SndAddr),
				ReceiverShard:     mb.ReceiverShardID,
				SenderShard:       mb.SenderShardID,
				GasPrice:          currentTx.GasPrice,
				GasLimit:          currentTx.GasLimit,
				Data:              hex.EncodeToString(currentTx.Data),
				Signature:         hex.EncodeToString(currentTx.Signature),
				Timestamp:         time.Duration(header.TimeStamp),
				Status:            mbTxStatus,
			})
		}
	}
	return bulks
}

// SaveMetaBlock will build
func (ei *elasticIndexer) SaveMetaBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header) {
	// Save Miniblocks
	// Save Block
	// Save Header
	if metaBlock == nil {
		fmt.Println("elasticsearch - no metaBlock")
		return
	}
	ei.saveMetaBlock(metaBlock, headerPool)
}

func (ei *elasticIndexer) saveMetaBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header) {
	var buff bytes.Buffer

	blocks := ei.buildBlocks(metaBlock, headerPool)

	for _, block := range blocks {

		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, block.Hash, "\n"))
		serializedBlock, err := json.Marshal(block)
		if err != nil {
			ei.logger.Warn("could not serialize block, will skip indexing: ", block.Hash)
			continue
		}
		// append a newline foreach element
		serializedBlock = append(serializedBlock, "\n"...)

		buff.Grow(len(meta) + len(serializedBlock))
		buff.Write(meta)
		buff.Write(serializedBlock)
	}

	res, err := ei.db.Bulk(bytes.NewReader(buff.Bytes()), ei.db.Bulk.WithIndex(blockIndex))
	buff.Reset()

	if err != nil {
		ei.logger.Warn("error indexing blocks")
	}
	if res.IsError() {
		fmt.Println(res.String())
		ei.logger.Warn("error from elasticsearch indexing blocks")
	}
}

// buildTransactionBulks creates bulks of maximum txBulkSize transactions to be indexed together
//  using the elasticsearch bulk API
func (ei *elasticIndexer) buildBlocks(metaBlock *block.MetaBlock, headerPool map[string]*block.Header) []*Block {

	blocks := make([]*Block, 0)

	//blockMarshal, _ := ei.marshalizer.Marshal(metaBlock)
	//blockHash := ei.hasher.Compute(string(blockMarshal))

	for _, shardData := range metaBlock.ShardInfo {
		headerHash := string(shardData.HeaderHash)
		headerRaw := *headerPool[headerHash]

		//TODO: add real size
		headerMarshal, _ := ei.marshalizer.Marshal(headerRaw)
		headerSize := len(headerMarshal)

		blocks = append(blocks, &Block{
			Hash:          headerHash,
			TxCount:       headerRaw.TxCount,
			Timestamp:     time.Duration(headerRaw.TimeStamp),
			StateRootHash: hex.EncodeToString(headerRaw.RootHash),
			Size:          int64(headerSize),
			ShardID:       shardData.ShardId,
			PubKeyBitmap:  hex.EncodeToString(headerRaw.PubKeysBitmap),
			Nonce:         headerRaw.Nonce,
			PrevHash:      hex.EncodeToString(headerRaw.PrevHash),
		})

	}

	return blocks
}
