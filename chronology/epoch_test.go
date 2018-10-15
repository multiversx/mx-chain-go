package chronology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEpoch(t *testing.T) {
	genesisTime := time.Now()
	index := 0
	epc := NewEpoch(index, genesisTime)

	epc.Print()

	assert.Equal(t, epc.index, index)
	assert.Equal(t, epc.genesisTime, genesisTime)
}
