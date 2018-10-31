package chronology_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/stretchr/testify/assert"
)

func TestEpoch(t *testing.T) {
	genesisTime := time.Now()
	index := 0
	epc := chronology.NewEpoch(index, genesisTime)

	epc.Print()

	assert.Equal(t, epc.Index(), index)
	assert.Equal(t, epc.GenesisTime(), genesisTime)
}
