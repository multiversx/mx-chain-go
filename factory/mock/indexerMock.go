package mock

import (
	indexerTypes "github.com/ElrondNetwork/elastic-indexer-go/types"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(args *indexerTypes.ArgsSaveBlockData)
}

// SaveBlock -
func (im *IndexerMock) SaveBlock(args *indexerTypes.ArgsSaveBlockData) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(args)
	}
}

// SetTxLogsProcessor will do nothing
func (im *IndexerMock) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// Close will do nothing
func (im *IndexerMock) Close() error {
	return nil
}

// SaveValidatorsRating --
func (im *IndexerMock) SaveValidatorsRating(_ string, _ []*indexerTypes.ValidatorRatingInfo) {

}

// SaveMetaBlock -
func (im *IndexerMock) SaveMetaBlock(_ data.HeaderHandler, _ []uint64) {
}

// UpdateTPS -
func (im *IndexerMock) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveRoundsInfo -
func (im *IndexerMock) SaveRoundsInfo(_ []*indexerTypes.RoundInfo) {
}

// SaveValidatorsPubKeys -
func (im *IndexerMock) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
	panic("implement me")
}

// RevertIndexedBlock -
func (im *IndexerMock) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveAccounts -
func (im *IndexerMock) SaveAccounts(_ uint64, _ []state.UserAccountHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerMock) IsInterfaceNil() bool {
	return im == nil
}

// IsNilIndexer -
func (im *IndexerMock) IsNilIndexer() bool {
	return false
}
