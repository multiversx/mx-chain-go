package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

// IndexerStub is a mock implementation fot the Indexer interface
type IndexerStub struct {
	SaveBlockCalled func(args *outportcore.ArgsSaveBlockData)
}

// SaveBlock -
func (im *IndexerStub) SaveBlock(args *outportcore.ArgsSaveBlockData) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(args)
	}
}

// Close will do nothing
func (im *IndexerStub) Close() error {
	return nil
}

// SetTxLogsProcessor will do nothing
func (im *IndexerStub) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// SaveRoundsInfo -
func (im *IndexerStub) SaveRoundsInfo(_ []*outportcore.RoundInfo) {
	panic("implement me")
}

// SaveValidatorsRating -
func (im *IndexerStub) SaveValidatorsRating(_ string, _ []*outportcore.ValidatorRatingInfo) {

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
