package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// ValidatorStatisticsProcessorStub -
type ValidatorStatisticsProcessorStub struct {
	ProcessCalled        func(info data.ValidatorInfoHandler) error
	IsInterfaceNilCalled func() bool
}

// Process -
func (pm *ValidatorStatisticsProcessorStub) Process(info data.ValidatorInfoHandler) error {
	if pm.ProcessCalled != nil {
		return pm.ProcessCalled(info)
	}

	return nil
}

// IsInterfaceNil -
func (pm *ValidatorStatisticsProcessorStub) IsInterfaceNil() bool {
	if pm.IsInterfaceNilCalled != nil {
		return pm.IsInterfaceNilCalled()
	}
	return false
}
