package persistor

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestPrepareMetricMaps(t *testing.T) {
	t.Parallel()

	metricMap := map[string]interface{}{
		core.MetricCountConsensus:               float64(1),
		core.MetricCountConsensusAcceptedBlocks: float64(2),
		core.MetricCountAcceptedBlocks:          float64(3),
		core.MetricCountLeader:                  float64(4),
		core.MetricNumProcessedTxs:              float64(5),
		core.MetricNumShardHeadersProcessed:     float64(6),
	}

	mapUint64Values, mapStringValues := prepareMetricMaps(metricMap)

	assert.Equal(t, uint64(1), mapUint64Values[core.MetricCountConsensus])
	assert.Equal(t, uint64(2), mapUint64Values[core.MetricCountConsensusAcceptedBlocks])
	assert.Equal(t, uint64(3), mapUint64Values[core.MetricCountAcceptedBlocks])
	assert.Equal(t, uint64(4), mapUint64Values[core.MetricCountLeader])
	assert.Equal(t, uint64(5), mapUint64Values[core.MetricNumProcessedTxs])
	assert.Equal(t, uint64(6), mapUint64Values[core.MetricNumShardHeadersProcessed])
	assert.Equal(t, map[string]string{}, mapStringValues)

}
