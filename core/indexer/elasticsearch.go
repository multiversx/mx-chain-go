package indexer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
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
const validatorsIndex = "validators"
const roundIndex = "rounds"

const metachainTpsDocID = "meta"
const shardTpsDocIDPrefix = "shard"

const badRequest = 400

var log = logger.GetOrCreate("core/indexer")

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
	options          *Options
	isNilIndexer     bool
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
	options *Options,
) (Indexer, error) {

	err := checkElasticSearchParams(
		url,
		shardCoordinator,
		marshalizer,
		hasher,
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
		options,
		false,
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

	err = indexer.checkAndCreateIndex(validatorsIndex, nil)
	if err != nil {
		return nil, err
	}

	err = indexer.checkAndCreateIndex(roundIndex, timestampMapping())
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
) error {
	if url == "" {
		return core.ErrNilUrl
	}
	if coordinator == nil || coordinator.IsInterfaceNil() {
		return core.ErrNilCoordinator
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return core.ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return core.ErrNilHasher
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

		log.Warn("indexer: resource already exists", "error", res.String())
		return ErrCannotCreateIndex
	}

	return nil
}

// SaveBlock will build
func (ei *elasticIndexer) SaveBlock(
	bodyHandler data.BodyHandler,
	headerhandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
) {

	if headerhandler == nil || headerhandler.IsInterfaceNil() {
		log.Debug("indexer: no header", "error", ErrNoHeader.Error())
		return
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		log.Debug("indexer", "error", ErrBodyTypeAssertion.Error())
		return
	}

	go ei.saveHeader(headerhandler, signersIndexes)

	if len(body) == 0 {
		log.Debug("indexer", "error", ErrNoMiniblocks.Error())
		return
	}

	if ei.options.TxIndexingEnabled {
		go ei.saveTransactions(body, headerhandler, txPool)
	}
}

// SaveMetaBlock will index a meta block in elastic search
func (ei *elasticIndexer) SaveMetaBlock(header data.HeaderHandler, signersIndexes []uint64) {
	if header == nil || header.IsInterfaceNil() {
		log.Debug("indexer: nil header", "error", ErrNoHeader.Error())
		return
	}

	go ei.saveHeader(header, signersIndexes)
}

// SaveRoundInfo will save data about a round on elastic search
func (ei *elasticIndexer) SaveRoundInfo(roundInfo RoundInfo) {
	var buff bytes.Buffer

	marshalizedRoundInfo, err := ei.marshalizer.Marshal(roundInfo)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal signers indexes")
		return
	}

	buff.Grow(len(marshalizedRoundInfo))
	buff.Write(marshalizedRoundInfo)

	req := esapi.IndexRequest{
		Index:      roundIndex,
		DocumentID: strconv.FormatUint(uint64(roundInfo.ShardId), 10) + "_" + strconv.FormatUint(roundInfo.Index, 10),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), ei.db)
	if err != nil {
		log.Warn("indexer: can not index round info", "error", err.Error())
		return
	}

	defer closeESResponseBody(res)

	if res.IsError() {
		log.Warn("indexer", "error", res.String())
	}
}

//SaveValidatorsPubKeys will send all validators public keys to elastic search
func (ei *elasticIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte) {
	valPubKeys := make(map[uint32][]string, 0)
	for shardId, shardPubKeys := range validatorsPubKeys {
		for _, pubKey := range shardPubKeys {
			valPubKeys[shardId] = append(valPubKeys[shardId], hex.EncodeToString(pubKey))
		}
		go func(id uint32, publicKeys []string) {
			ei.saveShardValidatorsPubKeys(id, publicKeys)
		}(shardId, valPubKeys[shardId])
	}
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ei *elasticIndexer) IsNilIndexer() bool {
	return ei.isNilIndexer
}

func (ei *elasticIndexer) saveShardValidatorsPubKeys(shardId uint32, shardValidatorsPubKeys []string) {
	var buff bytes.Buffer

	shardValPubKeys := ValidatorsPublicKeys{PublicKeys: shardValidatorsPubKeys}
	marshalizedValidatorPubKeys, err := ei.marshalizer.Marshal(shardValPubKeys)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators public keys")
		return
	}

	buff.Grow(len(marshalizedValidatorPubKeys))
	buff.Write(marshalizedValidatorPubKeys)

	req := esapi.IndexRequest{
		Index:      validatorsIndex,
		DocumentID: strconv.FormatUint(uint64(shardId), 10),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), ei.db)
	if err != nil {
		log.Warn("indexer: can not index validators pubkey", "error", err.Error())
		return
	}
	log.Debug("indexer: response validators pubkey", "msg", res.String())

	defer closeESResponseBody(res)

	if res.IsError() {
		log.Warn("indexer", "error", res.String())
	}
}

func (ei *elasticIndexer) getSerializedElasticBlockAndHeaderHash(header data.HeaderHandler, signersIndexes []uint64) ([]byte, []byte) {
	h, err := ei.marshalizer.Marshal(header)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal header")
		return nil, nil
	}

	headerHash := ei.hasher.Compute(string(h))
	elasticBlock := Block{
		Nonce:         header.GetNonce(),
		Round:         header.GetRound(),
		ShardID:       header.GetShardID(),
		Hash:          hex.EncodeToString(headerHash),
		Proposer:      signersIndexes[0],
		Validators:    signersIndexes,
		PubKeyBitmap:  hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:          int64(len(h)),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		TxCount:       header.GetTxCount(),
		StateRootHash: hex.EncodeToString(header.GetRootHash()),
		PrevHash:      hex.EncodeToString(header.GetPrevHash()),
	}

	serializedBlock, err := json.Marshal(elasticBlock)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal elastic header")
		return nil, nil
	}

	return serializedBlock, headerHash
}

func (ei *elasticIndexer) saveHeader(header data.HeaderHandler, signersIndexes []uint64) {
	var buff bytes.Buffer

	serializedBlock, headerHash := ei.getSerializedElasticBlockAndHeaderHash(header, signersIndexes)

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
		log.Warn("indexer: could not index block header", "error", err.Error())
		return
	}

	defer closeESResponseBody(res)

	if res.IsError() {
		log.Warn("indexer", "error", res.String())
	}
}

func (ei *elasticIndexer) serializeBulkTx(bulk []*Transaction) bytes.Buffer {
	var buff bytes.Buffer
	for _, tx := range bulk {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, tx.Hash, "_doc", "\n"))
		serializedTx, err := json.Marshal(tx)
		if err != nil {
			log.Debug("indexer: marshal",
				"error", "could not serialize transaction, will skip indexing",
				"tx hash", tx.Hash)
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
			log.Warn("indexer", "error", "indexing bulk of transactions")
			continue
		}
		if res.IsError() {
			log.Warn("indexer", "error", res.String())
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
			log.Debug("indexer: marshal", "error", "could not marshal miniblock")
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
				log.Debug("indexer: elasticsearch could not find tx hash in pool")
				continue
			}

			currentTx := getTransactionByType(currentTxHandler, txHash, mbHash, blockHash, mb, header, mbTxStatus)
			if currentTx == nil {
				log.Debug("indexer: elasticsearch found tx in pool but of wrong type")
				continue
			}

			bulks[currentBulk] = append(bulks[currentBulk], currentTx)
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
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this shard")
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (ei *elasticIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if tpsBenchmark == nil {
		log.Debug("indexer: update tps called, but the tpsBenchmark is nil")
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
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this round")
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
			log.Warn("indexer: error indexing tps information")
			continue
		}
		if res.IsError() {
			log.Warn("indexer", "error", res.String())
		}

		closeESResponseBody(res)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticIndexer) IsInterfaceNil() bool {
	if ei == nil {
		return true
	}
	return false
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

func getTransactionByType(
	tx data.TransactionHandler,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *Transaction {
	currentTx, ok := tx.(*transaction.Transaction)
	if ok && currentTx != nil {
		return buildTransaction(currentTx, txHash, mbHash, blockHash, mb, header, txStatus)
	}

	currentSc, ok := tx.(*smartContractResult.SmartContractResult)
	if ok && currentSc != nil {
		return buildSmartContractResult(currentSc, txHash, mbHash, blockHash, mb, header)
	}

	currentReward, ok := tx.(*rewardTx.RewardTx)
	if ok && currentReward != nil {
		return buildRewardTransaction(currentReward, txHash, mbHash, blockHash, mb, header)
	}

	return nil
}

func buildTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *Transaction {
	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         tx.Nonce,
		Round:         header.GetRound(),
		Value:         tx.Value.String(),
		Receiver:      hex.EncodeToString(tx.RcvAddr),
		Sender:        hex.EncodeToString(tx.SndAddr),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      tx.GasPrice,
		GasLimit:      tx.GasLimit,
		Data:          tx.Data,
		Signature:     hex.EncodeToString(tx.Signature),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
	}
}

func buildSmartContractResult(
	scr *smartContractResult.SmartContractResult,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
) *Transaction {
	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         scr.Nonce,
		Round:         header.GetRound(),
		Value:         scr.Value.String(),
		Receiver:      hex.EncodeToString(scr.RcvAddr),
		Sender:        hex.EncodeToString(scr.SndAddr),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          scr.Data,
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        "Success",
	}
}

func buildRewardTransaction(
	rTx *rewardTx.RewardTx,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
) *Transaction {

	shardIdStr := fmt.Sprintf("Shard%d", rTx.ShardId)

	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Receiver:      hex.EncodeToString(rTx.RcvAddr),
		Sender:        shardIdStr,
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          []byte(""),
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        "Success",
	}
}
