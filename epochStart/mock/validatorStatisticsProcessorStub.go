package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ValidatorStatisticsProcessorStub -
type ValidatorStatisticsProcessorStub struct {
	ProcessCalled        func(validatorInfo data.ShardValidatorInfoHandler) error
	CommitCalled         func() ([]byte, error)
	IsInterfaceNilCalled func() bool
}

// Process -
func (pm *ValidatorStatisticsProcessorStub) Process(validatorInfo data.ShardValidatorInfoHandler) error {
	if pm.ProcessCalled != nil {
		return pm.ProcessCalled(validatorInfo)
	}

	return nil
}

// Commit -
func (pm *ValidatorStatisticsProcessorStub) Commit() ([]byte, error) {
	if pm.CommitCalled != nil {
		return pm.CommitCalled()
	}

	return nil, nil
}

// IsInterfaceNil -
func (pm *ValidatorStatisticsProcessorStub) IsInterfaceNil() bool {
	if pm.IsInterfaceNilCalled != nil {
		return pm.IsInterfaceNilCalled()
	}
	return false
}
