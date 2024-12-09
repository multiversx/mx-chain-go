package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignEnableEpochsHandler struct {
	*enableEpochsHandler
	sovereignEnableEpochsConfig              config.SovereignEnableEpochs
	sovereignChainSpecificEnableEpochsConfig config.SovereignChainSpecificEnableEpochs
}

// NewSovereignEnableEpochsHandler creates a new instance of sovereign enable epochs handler
func NewSovereignEnableEpochsHandler(
	enableEpochsConfig config.EnableEpochs,
	sovereignEpochConfig config.SovereignEpochConfig,
	epochNotifier process.EpochNotifier,
) (*sovereignEnableEpochsHandler, error) {
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	sovHandler := &sovereignEnableEpochsHandler{
		enableEpochsHandler: &enableEpochsHandler{
			enableEpochsConfig: enableEpochsConfig,
		},
		sovereignEnableEpochsConfig:              sovereignEpochConfig.SovereignEnableEpochs,
		sovereignChainSpecificEnableEpochsConfig: sovereignEpochConfig.SovereignChainSpecificEnableEpochs,
	}

	sovHandler.createAllFlagsMap()

	epochNotifier.RegisterNotifyHandler(sovHandler)

	return sovHandler, nil
}

func (sovHandler *sovereignEnableEpochsHandler) createAllFlagsMap() {
	sovHandler.enableEpochsHandler.createAllFlagsMap()
	sovHandler.addSovereignFlags()
	sovHandler.addSovereignChainSpecificFlags()
}

func (sovHandler *sovereignEnableEpochsHandler) addSovereignFlags() {
	// follow the implementation from enableEpochsHandler.go to add a new sovereign flag
	// add the new flags in sovHandler.enableEpochsHandler.allFlagsDefined map
}

func (sovHandler *sovereignEnableEpochsHandler) addSovereignChainSpecificFlags() {
	// follow the implementation from enableEpochsHandler.go to add a new sovereign flag
	// add the new flags in sovHandler.enableEpochsHandler.allFlagsDefined map
}

// IsInterfaceNil returns true if there is no value under the interface
func (sovHandler *sovereignEnableEpochsHandler) IsInterfaceNil() bool {
	return sovHandler == nil
}
