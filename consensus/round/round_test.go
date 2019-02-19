package round_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/stretchr/testify/assert"
)

const roundTimeDuration = time.Duration(10 * time.Millisecond)

func TestRound_NewRoundShouldReturnNotNilRoundObject(t *testing.T) {
	genesisTime := time.Now()

	rnd := round.NewRound(genesisTime, genesisTime, roundTimeDuration)

	assert.NotNil(t, rnd)
}

func TestRound_UpdateRoundShouldNotChangeAnything(t *testing.T) {
	genesisTime := time.Now()

	rnd := round.NewRound(genesisTime, genesisTime, roundTimeDuration)
	oldIndex := rnd.Index()
	oldTimeStamp := rnd.TimeStamp()

	rnd.UpdateRound(genesisTime, genesisTime)

	newIndex := rnd.Index()
	newTimeStamp := rnd.TimeStamp()

	assert.Equal(t, oldIndex, newIndex)
	assert.Equal(t, oldTimeStamp, newTimeStamp)

}

func TestRound_UpdateRoundShouldAdvanceOneRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := round.NewRound(genesisTime, genesisTime, roundTimeDuration)
	oldIndex := rnd.Index()
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))
	newIndex := rnd.Index()

	assert.Equal(t, oldIndex, newIndex-1)
}

func TestRound_IndexShouldReturnFirstIndex(t *testing.T) {
	genesisTime := time.Now()

	rnd := round.NewRound(genesisTime, genesisTime, roundTimeDuration)
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration/2))
	index := rnd.Index()

	assert.Equal(t, int32(0), index)
}

func TestRound_TimeStampShouldReturnTimeStampOfTheNextRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := round.NewRound(genesisTime, genesisTime, roundTimeDuration)
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration+roundTimeDuration/2))
	timeStamp := rnd.TimeStamp()

	assert.Equal(t, genesisTime.Add(roundTimeDuration), timeStamp)
}

func TestRound_TimeDurationShouldReturnTheDurationOfOneRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := round.NewRound(genesisTime, genesisTime, roundTimeDuration)
	timeDuration := rnd.TimeDuration()

	assert.Equal(t, roundTimeDuration, timeDuration)
}
