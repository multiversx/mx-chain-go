package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
)

// DriverStub -
type DriverStub struct {
	SaveBlockCalled             func(args *outportcore.ArgsSaveBlockData) error
	RevertBlockCalled           func(header data.HeaderHandler, body data.BodyHandler) error
	SaveRoundsInfoCalled        func(roundsInfos []*outportcore.RoundInfo) error
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error
	SaveValidatorsRatingCalled  func(indexID string, infoRating []*outportcore.ValidatorRatingInfo) error
	SaveAccountsCalled          func(timestamp uint64, acc map[string]*outportcore.AlteredAccount) error
	FinalizedBlockCalled        func(headerHash []byte) error
	CloseCalled                 func() error
}

// SaveBlock -
func (d *DriverStub) SaveBlock(args *outportcore.ArgsSaveBlockData) error {
	if d.SaveBlockCalled != nil {
		return d.SaveBlockCalled(args)
	}

	return nil
}

// RevertIndexedBlock -
func (d *DriverStub) RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) error {
	if d.RevertBlockCalled != nil {
		return d.RevertBlockCalled(header, body)
	}

	return nil
}

// SaveRoundsInfo -
func (d *DriverStub) SaveRoundsInfo(roundsInfos []*outportcore.RoundInfo) error {
	if d.SaveRoundsInfoCalled != nil {
		return d.SaveRoundsInfoCalled(roundsInfos)
	}

	return nil
}

// SaveValidatorsPubKeys -
func (d *DriverStub) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
	if d.SaveValidatorsPubKeysCalled != nil {
		return d.SaveValidatorsPubKeysCalled(validatorsPubKeys, epoch)
	}

	return nil
}

// SaveValidatorsRating -
func (d *DriverStub) SaveValidatorsRating(indexID string, infoRating []*outportcore.ValidatorRatingInfo) error {
	if d.SaveValidatorsRatingCalled != nil {
		return d.SaveValidatorsRatingCalled(indexID, infoRating)
	}

	return nil
}

// SaveAccounts -
func (d *DriverStub) SaveAccounts(timestamp uint64, acc map[string]*outportcore.AlteredAccount, _ uint32) error {
	if d.SaveAccountsCalled != nil {
		return d.SaveAccountsCalled(timestamp, acc)
	}

	return nil
}

// FinalizedBlock -
func (d *DriverStub) FinalizedBlock(headerHash []byte) error {
	if d.FinalizedBlockCalled != nil {
		return d.FinalizedBlockCalled(headerHash)
	}

	return nil
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
