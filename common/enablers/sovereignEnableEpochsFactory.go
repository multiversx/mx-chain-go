package enablers

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignEnableEpochsFactory struct{}

// NewSovereignEnableEpochsFactory creates an enable epochs factory for sovereign chain
func NewSovereignEnableEpochsFactory() EnableEpochsFactory {
	return &sovereignEnableEpochsFactory{}
}

// CreateEnableEpochsHandler creates an enable epochs handler for sovereign chain
func (seef *sovereignEnableEpochsFactory) CreateEnableEpochsHandler(epochConfig config.EpochConfig, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error) {
	return NewSovereignEnableEpochsHandler(epochConfig.EnableEpochs, epochConfig.SovereignEnableEpochs, epochConfig.SovereignChainSpecificEnableEpochs,
		epochNotifier)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (seef *sovereignEnableEpochsFactory) IsInterfaceNil() bool {
	return seef == nil
}
