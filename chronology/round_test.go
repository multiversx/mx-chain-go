package chronology_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	rnd.Print()

	assert.Equal(t, rnd.Index(), int32(0))
	assert.Equal(t, rnd.TimeStamp(), genesisTime)
	assert.Equal(t, rnd.TimeDuration(), roundTimeDuration)

	currentTime = currentTime.Add(roundTimeDuration)

	rnd.UpdateRound(genesisTime, currentTime)

	assert.Equal(t, rnd.Index(), int32(1))
	assert.Equal(t, rnd.TimeStamp(), genesisTime.Add(roundTimeDuration))
	assert.Equal(t, rnd.TimeDuration(), roundTimeDuration)
}
