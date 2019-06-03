package indexer

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type ElasticIndexer struct {
	elasticIndexer
}

func NewTestElasticIndexer(url string, shardCoordinator sharding.Coordinator, marshalizer marshal.Marshalizer,
	hasher hashing.Hasher, logger *logger.Logger) ElasticIndexer {
	return ElasticIndexer{elasticIndexer{nil, shardCoordinator, marshalizer, hasher, logger}}
}

func (ei *ElasticIndexer) GetSerializedElasticBlockAndHeaderHash(header *block.Header) ([]byte, []byte) {
	return ei.getSerializedElasticBlockAndHeaderHash(header)
}

func (ei *ElasticIndexer) BuildTransactionBulks(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) [][]Transaction {
	return ei.buildTransactionBulks(body, header, txPool)
}

func (ei *ElasticIndexer) SerializeBulkTx(bulk []Transaction) bytes.Buffer {
	return ei.serializeBulkTx(bulk)
}
