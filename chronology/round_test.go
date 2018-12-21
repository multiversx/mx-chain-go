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
	rnd.UpdateRound(genesisTime, time.Now())
	newIndex := rnd.Index()

	assert.Equal(t, oldIndex, newIndex)
}

func TestUpdateRound_ShouldAdvanceOneRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	oldIndex := rnd.Index()
	time.Sleep(roundTimeDuration)
	rnd.UpdateRound(genesisTime, time.Now())
	newIndex := rnd.Index()

	assert.Equal(t, oldIndex, newIndex-1)
}

func TestIndex_ShouldReturnFirstIndex(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	time.Sleep(roundTimeDuration / 2)
	rnd.UpdateRound(genesisTime, time.Now())
	index := rnd.Index()

	assert.Equal(t, int32(0), index)
}

func TestTimeStamp_ShouldReturnTimeStampOfTheNextRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	time.Sleep(roundTimeDuration + roundTimeDuration/2)
	rnd.UpdateRound(genesisTime, time.Now())
	timeStamp := rnd.TimeStamp()

	assert.Equal(t, genesisTime.Add(roundTimeDuration), timeStamp)
}

func TestTimeDuration_ShouldReturnTheDurationOfOneRound(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)
	timeDuration := rnd.TimeDuration()

	assert.Equal(t, roundTimeDuration, timeDuration)
}

func TestPrint_ShouldPrintRoundObject(t *testing.T) {
	genesisTime := time.Now()

	rnd := chronology.NewRound(genesisTime, genesisTime, roundTimeDuration)

	rnd.Print()
}
