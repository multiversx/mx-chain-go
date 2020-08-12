package outport

import (
	"os"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
)

// CreateOutportDriver creates an outport driver
func CreateOutportDriver(config config.OutportConfig, txCoordinator TransactionCoordinator, logsProcessor TransactionLogProcessor) (Driver, error) {
	if !config.Enabled {
		log.Debug("Outport not enabled, will create a DisabledOutportDriver")
		return NewDisabledOutportDriver(), nil
	}

	log.Debug("Outport enabled, will create an OutputDriver")

	messagesMarshalizerKind := marshaling.ParseKind(config.MessagesMarshalizer)
	messagesMarshalizer := marshaling.CreateMarshalizer(messagesMarshalizerKind)
	namedPipe, err := os.OpenFile(config.NamedPipe, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	sender := NewSender(namedPipe, messagesMarshalizer)
	return newOutportDriver(config, txCoordinator, logsProcessor, sender)
}
