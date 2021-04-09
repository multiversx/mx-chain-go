package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// GetLastFinalizedHeaderHandlers returns the last finalized header handlers
func (es *EpochStart) GetLastFinalizedHeaderHandlers() []data.EpochStartShardDataHandler {
	if es == nil {
		return nil
	}

	epochStartShardDataHandlers := make([]data.EpochStartShardDataHandler, len(es.LastFinalizedHeaders))
	for i := range es.LastFinalizedHeaders {
		epochStartShardDataHandlers[i] = &es.LastFinalizedHeaders[i]
	}

	return epochStartShardDataHandlers
}

// GetEconomicsHandler  returns the economics handler
func (es *EpochStart) GetEconomicsHandler() data.EconomicsHandler {
	if es == nil {
		return nil
	}

	return &es.Economics
}

// SetLastFinalizedHeaders sets the last finalized header
func (es *EpochStart) SetLastFinalizedHeaders(epochStartDataHandlers []data.EpochStartShardDataHandler) error {
	if es == nil {
		return data.ErrNilPointerReceiver
	}

	epochStartData := make([]EpochStartShardData, len(epochStartDataHandlers))
	for i := range epochStartDataHandlers {
		shardData, ok := epochStartDataHandlers[i].(*EpochStartShardData)
		if !ok {
			return data.ErrInvalidTypeAssertion
		}
		if shardData == nil {
			return data.ErrNilPointerDereference
		}
		epochStartData[i] = *shardData
	}
	es.LastFinalizedHeaders = epochStartData
	return nil
}

// SetEconomics sets the economics data
func (es *EpochStart) SetEconomics(economicsHandler data.EconomicsHandler) error {
	if es == nil {
		return data.ErrNilPointerReceiver
	}
	ec, ok := economicsHandler.(*Economics)
	if !ok {
		return data.ErrInvalidTypeAssertion
	}
	if ec == nil {
		return data.ErrNilPointerDereference
	}
	es.Economics = *ec
	return nil
}
