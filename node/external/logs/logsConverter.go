package logs

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
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
		events[i] = &transaction.Events{
			Address:    converter.encodeAddress(event.Address),
			Identifier: string(event.Identifier),
			Topics:     event.Topics,
			Data:       event.Data,
		}
	}

	return &transaction.ApiLogs{
		Address: converter.encodeAddress(log.Address),
		Events:  events,
	}
}

func (converter *logsConverter) encodeAddress(pubkey []byte) string {
	return converter.pubKeyConverter.Encode(pubkey)
}
