package epoch_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/epoch"
	"github.com/stretchr/testify/assert"
)

func TestEpoch_NewEpochShouldWork(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()
	index := 0
	epc := epoch.NewEpoch(index, genesisTime)

	assert.Equal(t, epc.Index(), index)
	assert.Equal(t, epc.GenesisTime(), genesisTime)
}
