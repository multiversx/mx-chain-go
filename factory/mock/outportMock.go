package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

// OutportMock is a mock implementation fot the OutportHandler interface
type OutportMock struct {
	SaveBlockCalled func(args types.ArgsSaveBlocks)
}

// SaveBlock -
func (am *OutportMock) SaveBlock(args types.ArgsSaveBlocks) {
	if am.SaveBlockCalled != nil {
		am.SaveBlockCalled(args)
	}
}

// SaveValidatorsRating --
func (am *OutportMock) SaveValidatorsRating(_ string, _ []types.ValidatorRatingInfo) {

}

// SaveMetaBlock -
func (am *OutportMock) SaveMetaBlock(_ data.HeaderHandler, _ []uint64) {
}

// UpdateTPS -
func (am *OutportMock) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys -
func (am *OutportMock) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (am *OutportMock) IsInterfaceNil() bool {
	return am == nil
}

// HasDrivers -
func (am *OutportMock) HasDrivers() bool {
	return false
}

// RevertBlock -
func (am *OutportMock) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) {

}

// SaveAccounts -
func (am *OutportMock) SaveAccounts(_ []state.UserAccountHandler) {

}

// Close -
func (am *OutportMock) Close() error {
	return nil
}

// SaveRoundsInfo -
func (am *OutportMock) SaveRoundsInfo(_ []types.RoundInfo) {

}

// SubscribeDriver -
func (am *OutportMock) SubscribeDriver(_ drivers.Driver) error {
	return nil
}
