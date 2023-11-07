package logs

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type logsConverter struct {
	pubKeyConverter core.PubkeyConverter
}

func newLogsConverter(pubKeyConverter core.PubkeyConverter) *logsConverter {
	return &logsConverter{
		pubKeyConverter: pubKeyConverter,
	}
}

func (converter *logsConverter) txLogToApiResource(logKey []byte, log *transaction.Log) *transaction.ApiLogs {
	events := make([]*transaction.Events, len(log.Events))

	for i, event := range log.Events {
		eventAddress := converter.encodeAddress(event.Address)

		events[i] = &transaction.Events{
			Address:        eventAddress,
			Identifier:     string(event.Identifier),
			Topics:         event.Topics,
			Data:           event.Data,
			AdditionalData: event.AdditionalData,
		}
	}

	logAddress := converter.encodeAddress(log.Address)

	return &transaction.ApiLogs{
		Address: logAddress,
		Events:  events,
	}
}

func (converter *logsConverter) encodeAddress(pubkey []byte) string {
	return converter.pubKeyConverter.SilentEncode(pubkey, log)
}
