package enablers

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type enableEpochsFactory struct{}

func NewEnableEpochsFactory() EnableEpochsFactory {
	return &enableEpochsFactory{}
}

func (eef *enableEpochsFactory) CreateEnableEpochsHandler(epochConfig config.EpochConfig, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error) {
	return NewEnableEpochsHandler(epochConfig.EnableEpochs, epochNotifier)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (eef *enableEpochsFactory) IsInterfaceNil() bool {
	return eef == nil
}
