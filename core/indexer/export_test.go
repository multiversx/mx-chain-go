package indexer

import (
	"bytes"
	"io"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/elastic/go-elasticsearch/v7"
)

type ElasticIndexer struct {
	elasticIndexer
}

func NewTestElasticIndexer(
	url string,
	username string,
	password string,
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	logger *logger.Logger,
) ElasticIndexer {

	cfg := elasticsearch.Config{
		Addresses: []string{url},
		Username:  username,
		Password:  password,
	}

	es, _ := elasticsearch.NewClient(cfg)
	indexer := elasticIndexer{es, shardCoordinator,
		marshalizer, hasher, logger}

	return ElasticIndexer{indexer}
}

func (ei *ElasticIndexer) GetSerializedElasticBlockAndHeaderHash(header *block.Header) ([]byte, []byte) {
	return ei.getSerializedElasticBlockAndHeaderHash(header)
}

func (ei *ElasticIndexer) BuildTransactionBulks(
	body *block.Body,
	header *block.Header,
	txPool map[string]*transaction.Transaction,
) [][]*Transaction {
	return ei.buildTransactionBulks(body, header, txPool)
}

func (ei *ElasticIndexer) SerializeBulkTx(bulk []*Transaction) bytes.Buffer {
	return ei.serializeBulkTx(bulk)
}

func (ei *ElasticIndexer) SerializeShardInfo(shardInfo statistics.ShardStatistic) ([]byte, []byte) {
	return ei.serializeShardInfo(shardInfo)
}

func (ei *ElasticIndexer) CheckAndCreateIndex(index string, body io.Reader) error {
	return ei.checkAndCreateIndex(index, body)
}

func (ei *ElasticIndexer) CreateIndex(index string, body io.Reader) error {
	return ei.createIndex(index, body)
}
