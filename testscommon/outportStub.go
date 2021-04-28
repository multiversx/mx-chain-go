package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport"
)

// OutportStub is a mock implementation fot the OutportHandler interface
type OutportStub struct {
	SaveBlockCalled             func(args *indexer.ArgsSaveBlockData)
	SaveValidatorsRatingCalled  func(e string, k []*indexer.ValidatorRatingInfo)
	SaveValidatorsPubKeysCalled func(k map[uint32][][]byte, e uint32)
	HasDriversCalled            func() bool
}

// SaveBlock -
func (as *OutportStub) SaveBlock(args *indexer.ArgsSaveBlockData) {
	if as.SaveBlockCalled != nil {
		as.SaveBlockCalled(args)
	}
}

// SaveValidatorsRating --
func (as *OutportStub) SaveValidatorsRating(e string, k []*indexer.ValidatorRatingInfo) {
	if as.SaveValidatorsRatingCalled != nil {
		as.SaveValidatorsRatingCalled(e, k)
	}
}

// UpdateTPS -
func (as *OutportStub) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys -
func (as *OutportStub) SaveValidatorsPubKeys(k map[uint32][][]byte, e uint32) {
	if as.SaveValidatorsPubKeysCalled != nil {
		as.SaveValidatorsPubKeysCalled(k, e)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *OutportStub) IsInterfaceNil() bool {
	return as == nil
}

// HasDrivers -
func (as *OutportStub) HasDrivers() bool {
	if as.HasDriversCalled != nil {
		return as.HasDriversCalled()
	}
	return false
}

// RevertIndexedBlock -
func (as *OutportStub) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {

}

// SaveAccounts -
func (as *OutportStub) SaveAccounts(_ uint64, _ []state.UserAccountHandler) {

}

// Close -
func (as *OutportStub) Close() error {
	return nil
}

// SaveRoundsInfo -
func (as *OutportStub) SaveRoundsInfo(_ []*indexer.RoundInfo) {

}

// SubscribeDriver -
func (as *OutportStub) SubscribeDriver(_ outport.Driver) error {
	return nil
}
