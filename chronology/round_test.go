package chronology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)

	rnd.Print()

	assert.Equal(t, rnd.index, 0)
	assert.Equal(t, rnd.timeStamp, genesisTime)
	assert.Equal(t, rnd.timeDuration, ROUND_TIME_DURATION)

	currentTime = currentTime.Add(ROUND_TIME_DURATION)

	rnd.UpdateRound(genesisTime, currentTime)

	assert.Equal(t, rnd.index, 1)
	assert.Equal(t, rnd.timeStamp, genesisTime.Add(ROUND_TIME_DURATION))
	assert.Equal(t, rnd.timeDuration, ROUND_TIME_DURATION)
}
