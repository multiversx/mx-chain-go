package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// GetLastFinalizedHeaderHandlers -
func (m *EpochStart) GetLastFinalizedHeaderHandlers() []data.EpochStartShardDataHandler {
	if m == nil {
		return nil
	}

	epochStartShardDataHandlers := make([]data.EpochStartShardDataHandler, len(m.LastFinalizedHeaders))
	for i := range m.LastFinalizedHeaders {
		epochStartShardDataHandlers[i] = &m.LastFinalizedHeaders[i]
	}

	return epochStartShardDataHandlers
}

// GetEconomicsHandler -
func (m *EpochStart) GetEconomicsHandler() data.EconomicsHandler {
	if m == nil {
		return nil
	}

	return &m.Economics
}

// SetLastFinalizedHeaders -
func (m *EpochStart) SetLastFinalizedHeaders(epochStartDataHandlers []data.EpochStartShardDataHandler) {
	if m == nil {
		return
	}

	epochStartData := make([]EpochStartShardData, len(epochStartDataHandlers))
	for i := range epochStartDataHandlers {
		shardData, ok := epochStartDataHandlers[i].(*EpochStartShardData)
		if !ok {
			m.LastFinalizedHeaders = nil
			return
		}
		epochStartData[i] = *shardData
	}
	m.LastFinalizedHeaders = epochStartData
}

// SetEconomics -
func (m *EpochStart) SetEconomics(economicsHandler data.EconomicsHandler) {
	if m == nil {
		return
	}
	ec, ok := economicsHandler.(*Economics)
	if !ok {
		m.Economics = Economics{}
	}
	m.Economics = *ec
}
