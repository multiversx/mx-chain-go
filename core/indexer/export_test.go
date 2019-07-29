package indexer

import (
	"bytes"
	"io"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
	options *Options,
) ElasticIndexer {

	cfg := elasticsearch.Config{
		Addresses: []string{url},
		Username:  username,
		Password:  password,
	}

	es, _ := elasticsearch.NewClient(cfg)
	indexer := elasticIndexer{es, shardCoordinator,
		marshalizer, hasher, logger, options}

	return ElasticIndexer{indexer}
}

func (ei *ElasticIndexer) GetSerializedElasticBlockAndHeaderHash(header data.HeaderHandler) ([]byte, []byte) {
	return ei.getSerializedElasticBlockAndHeaderHash(header)
}

func (ei *ElasticIndexer) BuildTransactionBulks(
	body block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
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
