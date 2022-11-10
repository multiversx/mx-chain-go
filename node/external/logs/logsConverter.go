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

func (converter *logsConverter) txLogToApiResource(logKey []byte, log *transaction.Log) (*transaction.ApiLogs, error) {
	events := make([]*transaction.Events, len(log.Events))

	for i, event := range log.Events {
		encodedEventAddr, err := converter.encodeAddress(event.Address)
		if err != nil {
			return nil, err
		}
		events[i] = &transaction.Events{
			Address:    encodedEventAddr,
			Identifier: string(event.Identifier),
			Topics:     event.Topics,
			Data:       event.Data,
		}
	}

	encodedLogAddr, err := converter.encodeAddress(log.Address)
	if err != nil {
		return nil, err
	}

	return &transaction.ApiLogs{
		Address: encodedLogAddr,
		Events:  events,
	}, nil
}

func (converter *logsConverter) encodeAddress(pubkey []byte) (string, error) {
	return converter.pubKeyConverter.Encode(pubkey)
}
