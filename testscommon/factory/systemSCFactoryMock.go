package factory

import (
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// SysSCFactoryMock -
type SysSCFactoryMock struct {
	CreateSystemSCProcessorCalled func(args metachain.ArgsNewEpochStartSystemSCProcessing) (process.EpochStartSystemSCProcessor, error)
}

// CreateSystemSCProcessor -
func (mock *SysSCFactoryMock) CreateSystemSCProcessor(args metachain.ArgsNewEpochStartSystemSCProcessing) (process.EpochStartSystemSCProcessor, error) {
	if mock.CreateSystemSCProcessorCalled != nil {
		return mock.CreateSystemSCProcessorCalled(args)
	}

	return &testscommon.EpochStartSystemSCStub{}, nil
}

// IsInterfaceNil -
func (mock *SysSCFactoryMock) IsInterfaceNil() bool {
	return mock == nil
}
