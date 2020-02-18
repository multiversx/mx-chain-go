package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body *block.Body, header *block.Header)
}

// SaveBlock -
func (im *IndexerMock) SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64) {
	panic("implement me")
}

// SaveMetaBlock -
func (im *IndexerMock) SaveMetaBlock(header data.HeaderHandler, signersIndexes []uint64) {
}

// UpdateTPS -
func (im *IndexerMock) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	panic("implement me")
}

// SaveRoundInfo -
func (im *IndexerMock) SaveRoundInfo(roundInfo indexer.RoundInfo) {
	panic("implement me")
}

// SaveValidatorsPubKeys -
func (im *IndexerMock) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerMock) IsInterfaceNil() bool {
	return im == nil
}

// IsNilIndexer -
func (im *IndexerMock) IsNilIndexer() bool {
	return false
}
