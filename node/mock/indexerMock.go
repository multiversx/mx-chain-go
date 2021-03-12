package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// IndexerStub is a mock implementation fot the Indexer interface
type IndexerStub struct {
	SaveBlockCalled func(args *indexer.ArgsSaveBlockData)
}

// SaveBlock -
func (im *IndexerStub) SaveBlock(args *indexer.ArgsSaveBlockData) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(args)
	}

	return
}

// Close will do nothing
func (im *IndexerStub) Close() error {
	return nil
}

// SetTxLogsProcessor will do nothing
func (im *IndexerStub) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// UpdateTPS -
func (im *IndexerStub) UpdateTPS(_ statistics.TPSBenchmark) {
	panic("implement me")
}

// SaveRoundsInfo -
func (im *IndexerStub) SaveRoundsInfo(_ []*indexer.RoundInfo) {
	panic("implement me")
}

// SaveValidatorsRating -
func (im *IndexerStub) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) {

}

// SaveValidatorsPubKeys -
func (im *IndexerStub) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
	panic("implement me")
}

// RevertIndexedBlock -
func (im *IndexerStub) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveAccounts -
func (im *IndexerStub) SaveAccounts(_ uint64, _ []state.UserAccountHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerStub) IsInterfaceNil() bool {
	return im == nil
}

// IsNilIndexer -
func (im *IndexerStub) IsNilIndexer() bool {
	return false
}
