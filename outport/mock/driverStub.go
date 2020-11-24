package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

// DriverStub -
type DriverStub struct {
	SaveBlockCalled             func(args types.ArgsSaveBlocks)
	RevertBlockCalled           func(header data.HeaderHandler, body data.BodyHandler)
	SaveRoundsInfoCalled        func(roundsInfos []types.RoundInfo)
	UpdateTPSCalled             func(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRatingCalled  func(indexID string, infoRating []types.ValidatorRatingInfo)
	SaveAccountsCalled          func(acc []state.UserAccountHandler)
	CloseCalled                 func() error
}

// SaveBlock -
func (d *DriverStub) SaveBlock(args types.ArgsSaveBlocks) {
	if d.SaveBlockCalled != nil {
		d.SaveBlockCalled(args)
	}
}

// RevertBlock -
func (d *DriverStub) RevertBlock(header data.HeaderHandler, body data.BodyHandler) {
	if d.RevertBlockCalled != nil {
		d.RevertBlockCalled(header, body)
	}
}

// SaveRoundsInfo -
func (d *DriverStub) SaveRoundsInfo(roundsInfos []types.RoundInfo) {
	if d.SaveRoundsInfoCalled != nil {
		d.SaveRoundsInfoCalled(roundsInfos)
	}
}

// UpdateTPS -
func (d *DriverStub) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if d.UpdateTPSCalled != nil {
		d.UpdateTPSCalled(tpsBenchmark)
	}
}

// SaveValidatorsPubKeys -
func (d *DriverStub) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	if d.SaveValidatorsPubKeysCalled != nil {
		d.SaveValidatorsPubKeysCalled(validatorsPubKeys, epoch)
	}
}

// SaveValidatorsRating -
func (d *DriverStub) SaveValidatorsRating(indexID string, infoRating []types.ValidatorRatingInfo) {
	if d.SaveValidatorsRatingCalled != nil {
		d.SaveValidatorsRatingCalled(indexID, infoRating)
	}
}

// SaveAccounts -
func (d *DriverStub) SaveAccounts(acc []state.UserAccountHandler) {
	if d.SaveAccountsCalled != nil {
		d.SaveAccountsCalled(acc)
	}
}

// Close -
func (d *DriverStub) Close() error {
	if d.CloseCalled != nil {
		return d.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (d *DriverStub) IsInterfaceNil() bool {
	return d == nil
}
