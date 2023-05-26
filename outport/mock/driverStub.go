package mock

import (
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// DriverStub -
type DriverStub struct {
	SaveBlockCalled                         func(outportBlock *outportcore.OutportBlock) error
	RevertIndexedBlockCalled                func(blockData *outportcore.BlockData) error
	SaveRoundsInfoCalled                    func(roundsInfos *outportcore.RoundsInfo) error
	SaveValidatorsPubKeysCalled             func(validatorsPubKeys *outportcore.ValidatorsPubKeys) error
	SaveValidatorsRatingCalled              func(validatorsRating *outportcore.ValidatorsRating) error
	SaveAccountsCalled                      func(accounts *outportcore.Accounts) error
	FinalizedBlockCalled                    func(finalizedBlock *outportcore.FinalizedBlock) error
	CloseCalled                             func() error
	RegisterHandlerForSettingsRequestCalled func(handlerFunction func()) error
	CurrentSettingsCalled                   func(config outportcore.OutportConfig) error
}

// SaveBlock -
func (d *DriverStub) SaveBlock(args *outportcore.OutportBlock) error {
	if d.SaveBlockCalled != nil {
		return d.SaveBlockCalled(args)
	}

	return nil
}

// RevertIndexedBlock -
func (d *DriverStub) RevertIndexedBlock(blockData *outportcore.BlockData) error {
	if d.RevertIndexedBlockCalled != nil {
		return d.RevertIndexedBlockCalled(blockData)
	}

	return nil
}

// SaveRoundsInfo -
func (d *DriverStub) SaveRoundsInfo(roundsInfos *outportcore.RoundsInfo) error {
	if d.SaveRoundsInfoCalled != nil {
		return d.SaveRoundsInfoCalled(roundsInfos)
	}

	return nil
}

// SaveValidatorsPubKeys -
func (d *DriverStub) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) error {
	if d.SaveValidatorsPubKeysCalled != nil {
		return d.SaveValidatorsPubKeysCalled(validatorsPubKeys)
	}

	return nil
}

// SaveValidatorsRating -
func (d *DriverStub) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) error {
	if d.SaveValidatorsRatingCalled != nil {
		return d.SaveValidatorsRatingCalled(validatorsRating)
	}

	return nil
}

// SaveAccounts -
func (d *DriverStub) SaveAccounts(accounts *outportcore.Accounts) error {
	if d.SaveAccountsCalled != nil {
		return d.SaveAccountsCalled(accounts)
	}

	return nil
}

// FinalizedBlock -
func (d *DriverStub) FinalizedBlock(finalizedBlock *outportcore.FinalizedBlock) error {
	if d.FinalizedBlockCalled != nil {
		return d.FinalizedBlockCalled(finalizedBlock)
	}

	return nil
}

// GetMarshaller -
func (d *DriverStub) GetMarshaller() marshal.Marshalizer {
	return testscommon.MarshalizerMock{}
}

// CurrentSettings -
func (d *DriverStub) CurrentSettings(config outportcore.OutportConfig) error {
	if d.CurrentSettingsCalled != nil {
		return d.CurrentSettingsCalled(config)
	}

	return nil
}

// RegisterHandlerForSettingsRequest -
func (d *DriverStub) RegisterHandlerForSettingsRequest(handlerFunction func()) error {
	if d.RegisterHandlerForSettingsRequestCalled != nil {
		return d.RegisterHandlerForSettingsRequestCalled(handlerFunction)
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
