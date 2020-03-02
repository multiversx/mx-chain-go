package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler)
}

// SaveBlock -
func (im *IndexerMock) SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, _ []uint64) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(body, header, txPool)
	}
}

// SaveMetaBlock -
func (im *IndexerMock) SaveMetaBlock(_ data.HeaderHandler, _ []uint64) {
}

// UpdateTPS -
func (im *IndexerMock) UpdateTPS(_ statistics.TPSBenchmark) {
	panic("implement me")
}

// SaveRoundInfo -
func (im *IndexerMock) SaveRoundInfo(_ indexer.RoundInfo) {
}

// SaveValidatorsPubKeys -
func (im *IndexerMock) SaveValidatorsPubKeys(_ map[uint32][][]byte) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerMock) IsInterfaceNil() bool {
	return im == nil
}

// IsNilIndexer -
func (im *IndexerMock) IsNilIndexer() bool {
	return true
}
