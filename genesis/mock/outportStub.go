package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

// OutportStub is a mock implementation fot the OutportHandler interface
type OutportStub struct {
	SaveBlockCalled func(args types.ArgsSaveBlocks)
}

// SaveBlock -
func (as *OutportStub) SaveBlock(args types.ArgsSaveBlocks) {
	if as.SaveBlockCalled != nil {
		as.SaveBlockCalled(args)
	}
}

// SaveValidatorsRating --
func (as *OutportStub) SaveValidatorsRating(_ string, _ []types.ValidatorRatingInfo) {

}

// SaveMetaBlock -
func (as *OutportStub) SaveMetaBlock(_ data.HeaderHandler, _ []uint64) {
}

// UpdateTPS -
func (as *OutportStub) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys -
func (as *OutportStub) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *OutportStub) IsInterfaceNil() bool {
	return as == nil
}

// HasDrivers -
func (as *OutportStub) HasDrivers() bool {
	return false
}

// RevertBlock -
func (as *OutportStub) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) {

}

// SaveAccounts -
func (as *OutportStub) SaveAccounts(_ []state.UserAccountHandler) {

}

// Close -
func (as *OutportStub) Close() error {
	return nil
}

// SaveRoundsInfo -
func (as *OutportStub) SaveRoundsInfo(_ []types.RoundInfo) {

}

// SubscribeDriver -
func (as *OutportStub) SubscribeDriver(_ drivers.Driver) error {
	return nil
}
