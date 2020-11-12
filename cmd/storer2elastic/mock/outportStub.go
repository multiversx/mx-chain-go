package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

// OutportStub -
type OutportStub struct {
	SaveBlockCalled             func(args types.ArgsSaveBlocks)
	SaveRoundsInfosCalled       func(roundsInfos []types.RoundInfo)
	UpdateTPSCalled             func(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRatingCalled  func(indexID string, infoRating []types.ValidatorRatingInfo)
}

// SaveBlock -
func (os *OutportStub) SaveBlock(args types.ArgsSaveBlocks) {
	if os.SaveBlockCalled != nil {
		os.SaveBlockCalled(args)
	}
}

// SaveRoundsInfo -
func (os *OutportStub) SaveRoundsInfo(roundsInfos []types.RoundInfo) {
	if os.SaveRoundsInfosCalled != nil {
		os.SaveRoundsInfosCalled(roundsInfos)
	}
}

// UpdateTPS -
func (os *OutportStub) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if os.UpdateTPSCalled != nil {
		os.UpdateTPSCalled(tpsBenchmark)
	}
}

// SaveValidatorsPubKeys -
func (os *OutportStub) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	if os.SaveValidatorsPubKeysCalled != nil {
		os.SaveValidatorsPubKeysCalled(validatorsPubKeys, epoch)
	}
}

// SaveValidatorsRating -
func (os *OutportStub) SaveValidatorsRating(indexID string, infoRating []types.ValidatorRatingInfo) {
	if os.SaveValidatorsRatingCalled != nil {
		os.SaveValidatorsRatingCalled(indexID, infoRating)
	}
}

// Close -
func (os *OutportStub) Close() error {
	return nil
}

// RevertBlock -
func (os *OutportStub) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveAccounts -
func (os *OutportStub) SaveAccounts(_ []state.UserAccountHandler) {
}

// IsInterfaceNil -
func (os *OutportStub) IsInterfaceNil() bool {
	return os == nil
}

// SubscribeDriver -
func (os *OutportStub) SubscribeDriver(_ drivers.Driver) error {
	return nil
}

// HasDrivers -
func (os OutportStub) HasDrivers() bool {
	return false
}
