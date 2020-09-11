package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body *block.Body, header *block.Header)
}

// SaveBlock -
func (im *IndexerMock) SaveBlock(_ data.BodyHandler, _ data.HeaderHandler, _ map[string]data.TransactionHandler, _ []uint64, _ []string) {
	panic("implement me")
}

// StopIndexing will do nothing
func (im *IndexerMock) StopIndexing() error {
	return nil
}

// SetTxLogsProcessor will do nothing
func (im *IndexerMock) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// UpdateTPS -
func (im *IndexerMock) UpdateTPS(_ statistics.TPSBenchmark) {
	panic("implement me")
}

// SaveRoundsInfo -
func (im *IndexerMock) SaveRoundsInfo(_ []workItems.RoundInfo) {
	panic("implement me")
}

// SaveValidatorsRating --
func (im *IndexerMock) SaveValidatorsRating(_ string, _ []workItems.ValidatorRatingInfo) {

}

// SaveValidatorsPubKeys -
func (im *IndexerMock) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
	panic("implement me")
}

// RevertIndexedBlock -
func (im *IndexerMock) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerMock) IsInterfaceNil() bool {
	return im == nil
}

// IsNilIndexer -
func (im *IndexerMock) IsNilIndexer() bool {
	return false
}
