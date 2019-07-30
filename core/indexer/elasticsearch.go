package indexer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gin-gonic/gin/json"
)

const txBulkSize = 1000
const txIndex = "transactions"
const blockIndex = "blocks"
const tpsIndex = "tps"

const metachainTpsDocID = "meta"
const shardTpsDocIDPrefix = "shard"

const badRequest = 400

// Options structure holds the indexer's configuration options
type Options struct {
	TxIndexingEnabled bool
}

//TODO refactor this and split in 3: glue code, interface and logic code
type elasticIndexer struct {
	db               *elasticsearch.Client
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	logger           *logger.Logger
	options          *Options
}

// NewElasticIndexer creates a new elasticIndexer where the server listens on the url, authentication for the server is
// using the username and password
func NewElasticIndexer(
	url string,
	username string,
	password string,
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	logger *logger.Logger,
	options *Options,
) (Indexer, error) {

	err := checkElasticSearchParams(
		url,
		shardCoordinator,
		marshalizer,
		hasher,
		logger,
	)
	if err != nil {
		return nil, err
	}

	cfg := elasticsearch.Config{
		Addresses: []string{url},
		Username:  username,
		Password:  password,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	indexer := &elasticIndexer{
		es,
		shardCoordinator,
		marshalizer,
		hasher,
		logger,
		options,
	}

	err = indexer.checkAndCreateIndex(blockIndex, timestampMapping())
	if err != nil {
		return nil, err
	}

	err = indexer.checkAndCreateIndex(txIndex, timestampMapping())
	if err != nil {
		return nil, err
	}

	err = indexer.checkAndCreateIndex(tpsIndex, nil)
	if err != nil {
		return nil, err
	}

	return indexer, nil
}

func checkElasticSearchParams(
	url string,
	coordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	logger *logger.Logger,
) error {
	if url == "" {
		return core.ErrNilUrl
	}
	if coordinator == nil {
		return core.ErrNilCoordinator
	}
	if marshalizer == nil {
		return core.ErrNilMarshalizer
	}
	if hasher == nil {
		return core.ErrNilHasher
	}
	if logger == nil {
		return core.ErrNilLogger
	}

	return nil
}

func (ei *elasticIndexer) checkAndCreateIndex(index string, body io.Reader) error {
	res, err := ei.db.Indices.Exists([]string{index})
	if err != nil {
		return err
	}

	defer closeESResponseBody(res)
	// Indices.Exists actually does a HEAD request to the elastic index.
	// A status code of 200 actually means the index exists so we
	//  don't need to do anything.
	if res.StatusCode == http.StatusOK {
		return nil
	}
	// A status code of 404 means the index does not exist so we create it
	if res.StatusCode == http.StatusNotFound {
		err = ei.createIndex(index, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticIndexer) createIndex(index string, body io.Reader) error {
	var err error
	var res *esapi.Response

	if body != nil {
		res, err = ei.db.Indices.Create(
			index,
			ei.db.Indices.Create.WithBody(body))
	} else {
		res, err = ei.db.Indices.Create(index)
	}

	defer closeESResponseBody(res)

	if err != nil {
		return err
	}

	if res.IsError() {
		// Resource already exists
		if res.StatusCode == badRequest {
			return nil
		}

		ei.logger.Warn(res.String())
		return ErrCannotCreateIndex
	}

	return nil
}

// SaveBlock will build
func (ei *elasticIndexer) SaveBlock(
	bodyHandler data.BodyHandler,
	headerhandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler) {

	if headerhandler == nil || headerhandler.IsInterfaceNil() {
		ei.logger.Warn(ErrNoHeader.Error())
		return
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		ei.logger.Warn(ErrBodyTypeAssertion.Error())
		return
	}

	go ei.saveHeader(headerhandler)

	if len(body) == 0 {
		ei.logger.Warn(ErrNoMiniblocks.Error())
		return
	}

	if ei.options.TxIndexingEnabled {
		go ei.saveTransactions(body, headerhandler, txPool)
	}
}

func (ei *elasticIndexer) getSerializedElasticBlockAndHeaderHash(header data.HeaderHandler) ([]byte, []byte) {
	h, err := ei.marshalizer.Marshal(header)
	if err != nil {
		ei.logger.Warn("could not marshal header")
		return nil, nil
	}

	headerHash := ei.hasher.Compute(string(h))
	elasticBlock := Block{
		Nonce:   header.GetNonce(),
		ShardID: header.GetShardID(),
		Hash:    hex.EncodeToString(headerHash),
		// TODO: We should add functionality for proposer and validators
		Proposer: hex.EncodeToString([]byte("mock proposer")),
		//Validators: "mock validators",
		PubKeyBitmap:  hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:          int64(len(h)),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		TxCount:       header.GetTxCount(),
		StateRootHash: hex.EncodeToString(header.GetRootHash()),
		PrevHash:      hex.EncodeToString(header.GetPrevHash()),
	}

	serializedBlock, err := json.Marshal(elasticBlock)
	if err != nil {
		ei.logger.Warn("could not marshal elastic header")
		return nil, nil
	}

	return serializedBlock, headerHash
}

func (ei *elasticIndexer) saveHeader(header data.HeaderHandler) {
	var buff bytes.Buffer

	serializedBlock, headerHash := ei.getSerializedElasticBlockAndHeaderHash(header)

	buff.Grow(len(serializedBlock))
	buff.Write(serializedBlock)

	req := esapi.IndexRequest{
		Index:      blockIndex,
		DocumentID: hex.EncodeToString(headerHash),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), ei.db)
	if err != nil {
		ei.logger.Warn(fmt.Sprintf("Could not index block header: %s", err))
		return
	}

	defer closeESResponseBody(res)

	if res.IsError() {
		ei.logger.Warn(res.String())
	}
}

func (ei *elasticIndexer) serializeBulkTx(bulk []*Transaction) bytes.Buffer {
	var buff bytes.Buffer
	for _, tx := range bulk {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, tx.Hash, "_doc", "\n"))
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

	return buff
}

func (ei *elasticIndexer) saveTransactions(
	body block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler) {
	bulks := ei.buildTransactionBulks(body, header, txPool)

	for _, bulk := range bulks {
		buff := ei.serializeBulkTx(bulk)
		res, err := ei.db.Bulk(bytes.NewReader(buff.Bytes()), ei.db.Bulk.WithIndex(txIndex))
		if err != nil {
			ei.logger.Warn("error indexing bulk of transactions")
			continue
		}
		if res.IsError() {
			ei.logger.Warn(res.String())
		}

		closeESResponseBody(res)
	}
}

// buildTransactionBulks creates bulks of maximum txBulkSize transactions to be indexed together
//  using the elasticsearch bulk API
func (ei *elasticIndexer) buildTransactionBulks(
	body block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
) [][]*Transaction {
	processedTxCount := 0
	bulks := make([][]*Transaction, (header.GetTxCount()/txBulkSize)+1)
	blockMarshal, _ := ei.marshalizer.Marshal(header)
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
			currentTxHandler, ok := txPool[string(txHash)]
			if !ok {
				ei.logger.Warn("elasticsearch could not find tx hash in pool")
				continue
			}

			currentTx, ok := currentTxHandler.(*transaction.Transaction)
			if !ok {
				ei.logger.Warn("elasticsearch found tx in pool but of wrong type")
				continue
			}

			bulks[currentBulk] = append(bulks[currentBulk], &Transaction{
				Hash:          hex.EncodeToString(txHash),
				MBHash:        hex.EncodeToString(mbHash),
				BlockHash:     hex.EncodeToString(blockHash),
				Nonce:         currentTx.Nonce,
				Value:         currentTx.Value,
				Receiver:      hex.EncodeToString(currentTx.RcvAddr),
				Sender:        hex.EncodeToString(currentTx.SndAddr),
				ReceiverShard: mb.ReceiverShardID,
				SenderShard:   mb.SenderShardID,
				GasPrice:      currentTx.GasPrice,
				GasLimit:      currentTx.GasLimit,
				Data:          currentTx.Data,
				Signature:     hex.EncodeToString(currentTx.Signature),
				Timestamp:     time.Duration(header.GetTimeStamp()),
				Status:        mbTxStatus,
			})
		}
	}

	return bulks
}

func (ei *elasticIndexer) serializeShardInfo(shardInfo statistics.ShardStatistic) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s%d", "_type" : "%s" } }%s`,
		shardTpsDocIDPrefix, shardInfo.ShardID(), tpsIndex, "\n"))

	bigTxCount := big.NewInt(int64(shardInfo.AverageBlockTxCount()))
	shardTPS := TPS{
		ShardID:               shardInfo.ShardID(),
		LiveTPS:               shardInfo.LiveTPS(),
		PeakTPS:               shardInfo.PeakTPS(),
		AverageTPS:            shardInfo.AverageTPS(),
		AverageBlockTxCount:   bigTxCount,
		CurrentBlockNonce:     shardInfo.CurrentBlockNonce(),
		LastBlockTxCount:      shardInfo.LastBlockTxCount(),
		TotalProcessedTxCount: shardInfo.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(shardTPS)
	if err != nil {
		ei.logger.Warn("could not serialize tps info, will skip indexing tps this shard")
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (ei *elasticIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if tpsBenchmark == nil {
		ei.logger.Warn("update tps called, but the tpsBenchmark is nil")
		return
	}

	var buff bytes.Buffer

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, metachainTpsDocID, tpsIndex, "\n"))
	generalInfo := TPS{
		LiveTPS:    tpsBenchmark.LiveTPS(),
		PeakTPS:    tpsBenchmark.PeakTPS(),
		NrOfShards: tpsBenchmark.NrOfShards(),
		// TODO: This value is still mocked, it should be removed if we cannot populate it correctly
		NrOfNodes:             100,
		BlockNumber:           tpsBenchmark.BlockNumber(),
		RoundNumber:           tpsBenchmark.RoundNumber(),
		RoundTime:             tpsBenchmark.RoundTime(),
		AverageBlockTxCount:   tpsBenchmark.AverageBlockTxCount(),
		LastBlockTxCount:      tpsBenchmark.LastBlockTxCount(),
		TotalProcessedTxCount: tpsBenchmark.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(generalInfo)
	if err != nil {
		ei.logger.Warn("could not serialize tps info, will skip indexing tps this round")
		return
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	buff.Write(meta)
	buff.Write(serializedInfo)

	for _, shardInfo := range tpsBenchmark.ShardStatistics() {
		serializedInfo, meta := ei.serializeShardInfo(shardInfo)
		if serializedInfo == nil {
			continue
		}

		buff.Grow(len(meta) + len(serializedInfo))
		buff.Write(meta)
		buff.Write(serializedInfo)

		res, err := ei.db.Bulk(bytes.NewReader(buff.Bytes()), ei.db.Bulk.WithIndex(tpsIndex))
		if err != nil {
			ei.logger.Warn("error indexing tps information")
			continue
		}
		if res.IsError() {
			fmt.Println(res.String())
			ei.logger.Warn("error from elasticsearch indexing tps information")
		}

		closeESResponseBody(res)
	}
}

func closeESResponseBody(res *esapi.Response) {
	if res == nil {
		return
	}
	if res.Body == nil {
		return
	}

	_ = res.Body.Close()
}

func timestampMapping() io.Reader {
	return strings.NewReader(
		`{
				"settings": {"index": {"sort.field": "timestamp", "sort.order": "desc"}},
				"mappings": {"_doc": {"properties": {"timestamp": {"type": "date"}}}}
			}`,
	)
}
