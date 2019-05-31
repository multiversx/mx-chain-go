package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body block.Body, header *block.Header)
}

func (im *IndexerMock) SaveBlock(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction) {
	panic("implement me")
}

func (im *IndexerMock) SaveMetaBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header) {
	panic("implement me")
}

func (im *IndexerMock) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	panic("implement me")
}
