package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/state"
)

// DriverStub -
type DriverStub struct {
	SaveBlockCalled             func(args *indexer.ArgsSaveBlockData)
	RevertBlockCalled           func(header data.HeaderHandler, body data.BodyHandler)
	SaveRoundsInfoCalled        func(roundsInfos []*indexer.RoundInfo)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRatingCalled  func(indexID string, infoRating []*indexer.ValidatorRatingInfo)
	SaveAccountsCalled          func(timestamp uint64, acc []state.UserAccountHandler)
	CloseCalled                 func() error
}

// SaveBlock -
func (d *DriverStub) SaveBlock(args *indexer.ArgsSaveBlockData) {
	if d.SaveBlockCalled != nil {
		d.SaveBlockCalled(args)
	}
}

// RevertIndexedBlock -
func (d *DriverStub) RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) {
	if d.RevertBlockCalled != nil {
		d.RevertBlockCalled(header, body)
	}
}

// SaveRoundsInfo -
func (d *DriverStub) SaveRoundsInfo(roundsInfos []*indexer.RoundInfo) {
	if d.SaveRoundsInfoCalled != nil {
		d.SaveRoundsInfoCalled(roundsInfos)
	}
}

// SaveValidatorsPubKeys -
func (d *DriverStub) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	if d.SaveValidatorsPubKeysCalled != nil {
		d.SaveValidatorsPubKeysCalled(validatorsPubKeys, epoch)
	}
}

// SaveValidatorsRating -
func (d *DriverStub) SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) {
	if d.SaveValidatorsRatingCalled != nil {
		d.SaveValidatorsRatingCalled(indexID, infoRating)
	}
}

// SaveAccounts -
func (d *DriverStub) SaveAccounts(timestamp uint64, acc []state.UserAccountHandler) {
	if d.SaveAccountsCalled != nil {
		d.SaveAccountsCalled(timestamp, acc)
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
