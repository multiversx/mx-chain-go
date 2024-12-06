package common

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/mock"
)

// EnableEpochsFactoryMock -
type EnableEpochsFactoryMock struct {
	CreateEnableEpochsHandlerCaller func(epochConfig config.EpochConfig, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error)
}

// CreateEnableEpochsHandler -
func (f *EnableEpochsFactoryMock) CreateEnableEpochsHandler(epochConfig config.EpochConfig, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error) {
	if f.CreateEnableEpochsHandlerCaller != nil {
		return f.CreateEnableEpochsHandlerCaller(epochConfig, epochNotifier)
	}

	return &mock.EnableEpochsHandlerMock{}, nil
}

// IsInterfaceNil -
func (f *EnableEpochsFactoryMock) IsInterfaceNil() bool {
	return f == nil
}
