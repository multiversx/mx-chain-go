package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core"
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
	sovereignEnableEpochsConfig config.SovereignEnableEpochs,
	sovereignChainSpecificEnableEpochsConfig config.SovereignChainSpecificEnableEpochs,
	epochNotifier process.EpochNotifier,
) (*sovereignEnableEpochsHandler, error) {
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	sovHandler := &sovereignEnableEpochsHandler{
		enableEpochsHandler: &enableEpochsHandler{
			enableEpochsConfig: enableEpochsConfig,
		},
		sovereignEnableEpochsConfig:              sovereignEnableEpochsConfig,
		sovereignChainSpecificEnableEpochsConfig: sovereignChainSpecificEnableEpochsConfig,
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
	sovereignFlags := map[core.EnableEpochFlag]flagHandler{}

	for key, value := range sovereignFlags {
		sovHandler.enableEpochsHandler.allFlagsDefined[key] = value
	}
}

func (sovHandler *sovereignEnableEpochsHandler) addSovereignChainSpecificFlags() {
	sovereignChainSpecificFlags := map[core.EnableEpochFlag]flagHandler{}

	for key, value := range sovereignChainSpecificFlags {
		sovHandler.enableEpochsHandler.allFlagsDefined[key] = value
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sovHandler *sovereignEnableEpochsHandler) IsInterfaceNil() bool {
	return sovHandler == nil
}
