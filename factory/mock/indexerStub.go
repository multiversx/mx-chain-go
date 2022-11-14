package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
)

// IndexerStub is a mock implementation fot the Indexer interface
type IndexerStub struct {
	SaveBlockCalled func(args *outport.ArgsSaveBlockData)
}

// SaveBlock -
func (im *IndexerStub) SaveBlock(args *outport.ArgsSaveBlockData) {
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
func (im *IndexerStub) SaveRoundsInfo(_ []*outport.RoundInfo) {
	panic("implement me")
}

// SaveValidatorsRating -
func (im *IndexerStub) SaveValidatorsRating(_ string, _ []*outport.ValidatorRatingInfo) {

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
