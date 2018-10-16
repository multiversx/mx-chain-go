package chronology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := NewRound(genesisTime, currentTime, roundTimeDuration)

	rnd.Print()

	assert.Equal(t, rnd.index, 0)
	assert.Equal(t, rnd.timeStamp, genesisTime)
	assert.Equal(t, rnd.timeDuration, roundTimeDuration)

	currentTime = currentTime.Add(roundTimeDuration)

	rnd.UpdateRound(genesisTime, currentTime)

	assert.Equal(t, rnd.index, 1)
	assert.Equal(t, rnd.timeStamp, genesisTime.Add(roundTimeDuration))
	assert.Equal(t, rnd.timeDuration, roundTimeDuration)
}
