package outport

import (
	"github.com/ElrondNetwork/elrond-go/config"
)

func CreateOutportDriver(config config.OutportConfig, txCoordinator TransactionCoordinator, logsProcessor TransactionLogProcessor) (Driver, error) {
	if !config.Enabled {
		log.Debug("Outport not enabled, will create a DisabledOutportDriver")
		return NewDisabledOutportDriver(), nil
	}

	// serializer received

	log.Debug("Outport enabled, will create an OutputDriver")
	return NewOutportDriver(config, txCoordinator, logsProcessor)
}
