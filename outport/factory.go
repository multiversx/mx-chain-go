package outport

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
)

func CreateOutportDriver(config config.OutportConfig, txCoordinator TransactionCoordinator, logsProcessor TransactionLogProcessor) (Driver, error) {
	if !config.Enabled {
		log.Debug("Outport not enabled, will create a DisabledOutportDriver")
		return NewDisabledOutportDriver(), nil
	}

	log.Debug("Outport enabled, will create an OutputDriver")

	messagesMarshalizerKind := marshaling.ParseKind(config.MessagesMarshalizer)
	messagesMarshalizer := marshaling.CreateMarshalizer(messagesMarshalizerKind)
	return NewOutportDriver(config, txCoordinator, logsProcessor, messagesMarshalizer)
}
