package enablers

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignEnableEpochsFactory struct {
	sovereignEpochConfig config.SovereignEpochConfig
}

// NewSovereignEnableEpochsFactory creates an enable epochs factory for sovereign chain
func NewSovereignEnableEpochsFactory(sovereignEpochConfig config.SovereignEpochConfig) EnableEpochsFactory {
	return &sovereignEnableEpochsFactory{
		sovereignEpochConfig: sovereignEpochConfig,
	}
}

// CreateEnableEpochsHandler creates an enable epochs handler for sovereign chain
func (seef *sovereignEnableEpochsFactory) CreateEnableEpochsHandler(enableEpochs config.EnableEpochs, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error) {
	return NewSovereignEnableEpochsHandler(enableEpochs, seef.sovereignEpochConfig, epochNotifier)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (seef *sovereignEnableEpochsFactory) IsInterfaceNil() bool {
	return seef == nil
}
