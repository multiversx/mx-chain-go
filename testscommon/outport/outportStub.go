package outport

import (
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport"
)

// OutportStub is a mock implementation fot the OutportHandler interface
type OutportStub struct {
	SaveBlockCalled             func(args *outportcore.OutportBlockWithHeaderAndBody) error
	SaveValidatorsRatingCalled  func(validatorsRating *outportcore.ValidatorsRating)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys *outportcore.ValidatorsPubKeys)
	HasDriversCalled            func() bool
}

// SaveBlock -
func (as *OutportStub) SaveBlock(args *outportcore.OutportBlockWithHeaderAndBody) error {
	if as.SaveBlockCalled != nil {
		return as.SaveBlockCalled(args)
	}

	return nil
}

// SaveValidatorsRating -
func (as *OutportStub) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) {
	if as.SaveValidatorsRatingCalled != nil {
		as.SaveValidatorsRatingCalled(validatorsRating)
	}
}

// SaveValidatorsPubKeys -
func (as *OutportStub) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) {
	if as.SaveValidatorsPubKeysCalled != nil {
		as.SaveValidatorsPubKeysCalled(validatorsPubKeys)
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
func (as *OutportStub) RevertIndexedBlock(_ *outportcore.HeaderDataWithBody) error {
	return nil
}

// SaveAccounts -
func (as *OutportStub) SaveAccounts(_ *outportcore.Accounts) {

}

// Close -
func (as *OutportStub) Close() error {
	return nil
}

// SaveRoundsInfo -
func (as *OutportStub) SaveRoundsInfo(_ *outportcore.RoundsInfo) {

}

// SubscribeDriver -
func (as *OutportStub) SubscribeDriver(_ outport.Driver) error {
	return nil
}

// FinalizedBlock -
func (as *OutportStub) FinalizedBlock(_ *outportcore.FinalizedBlock) {
}
