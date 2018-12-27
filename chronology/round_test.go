package chronology_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/stretchr/testify/assert"
)

func TestNewRound_ShouldReturnNotNilRoundObject(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)

	assert.NotNil(t, rnd)
}

func TestUpdateRound_ShouldNotChangeAnything(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	oldIndex := rnd.Index()
	oldTimeStamp := rnd.TimeStamp()

	rnd.UpdateRound(genesisTime, genesisTime)

	newIndex := rnd.Index()
	newTimeStamp := rnd.TimeStamp()

	assert.Equal(t, oldIndex, newIndex)
	assert.Equal(t, oldTimeStamp, newTimeStamp)
}

func TestUpdateRound_ShouldAdvanceOneRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	oldIndex := rnd.Index()
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration))
	newIndex := rnd.Index()

	assert.Equal(t, oldIndex, newIndex-1)
}

func TestIndex_ShouldReturnFirstIndex(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration/2))
	index := rnd.Index()

	assert.Equal(t, int32(0), index)
}

func TestTimeStamp_ShouldReturnTimeStampOfTheNextRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	rnd.UpdateRound(genesisTime, genesisTime.Add(roundTimeDuration+roundTimeDuration/2))
	timeStamp := rnd.TimeStamp()

	assert.Equal(t, genesisTime.Add(roundTimeDuration), timeStamp)
}

func TestTimeDuration_ShouldReturnTheDurationOfOneRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	timeDuration := rnd.TimeDuration()

	assert.Equal(t, roundTimeDuration, timeDuration)
}
