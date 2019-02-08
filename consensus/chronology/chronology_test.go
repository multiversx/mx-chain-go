package chronology_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology/mock"
	"github.com/stretchr/testify/assert"
)

func initSubroundHandlerMock() *mock.SubroundHandlerMock {
	srm := &mock.SubroundHandlerMock{}

	srm.CurrentCalled = func() int {
		return 0
	}

	srm.NextCalled = func() int {
		return 1
	}

	srm.DoWorkCalled = func(haveTimeInThisRound func() time.Duration) bool {
		return false
	}

	srm.NameCalled = func() string {
		return "(TEST)"
	}

	return srm
}

func TestChronology_NewChronologyNilRounderShouldFail(t *testing.T) {
	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, err := chronology.NewChronology(
		genesisTime,
		nil,
		syncTimerMock)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilRounder)
}

func TestChronology_NewChronologyNilSyncerShouldFail(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	genesisTime := time.Now()

	chr, err := chronology.NewChronology(
		genesisTime,
		rounderMock,
		nil)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilSyncTimer)
}

func TestChronology_NewChronologyShouldWork(t *testing.T) {
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, err := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	assert.NotNil(t, chr)
	assert.Nil(t, err)
}

func TestChronology_AddSubroundShouldWork(t *testing.T) {
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())

	assert.Equal(t, 3, len(chr.SubroundHandlers()))
}

func TestChronology_RemoveAllSubroundsShouldReturnEmptySubroundHandlersArray(t *testing.T) {
	rounderMock := &mock.RounderMock{}
	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())

	assert.Equal(t, 3, len(chr.SubroundHandlers()))

	chr.RemoveAllSubrounds()

	assert.Equal(t, 0, len(chr.SubroundHandlers()))
}

func TestChronology_StartRoundShouldReturnWhenRoundIndexIsNegative(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	rounderMock.IndexCalled = func() int32 {
		return -1
	}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()

	chr.AddSubround(srm)

	chr.SetSubroundId(0)

	chr.StartRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldReturnWhenLoadSubroundHandlerReturnsNil(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()

	chr.AddSubround(srm)

	chr.SetSubroundId(0)

	chr.StartRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldReturnWhenDoWorkReturnsFalse(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()

	chr.AddSubround(srm)

	chr.SetSubroundId(0)

	chr.StartRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldWork(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()

	srm.DoWorkCalled = func(haveTimeInThisRound func() time.Duration) bool {
		return true
	}

	chr.AddSubround(srm)

	chr.SetSubroundId(0)

	chr.StartRound()

	assert.Equal(t, srm.Next(), chr.SubroundId())
}

func TestChronology_UpdateRoundShouldInitRound(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	srm := initSubroundHandlerMock()

	chr.AddSubround(srm)

	chr.UpdateRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_LoadSubrounderShouldReturnNilWhenSubroundHandlerNotExists(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	assert.Nil(t, chr.LoadSubroundHandler(0))
}

func TestChronology_LoadSubrounderShouldReturnNilWhenIndexIsOutOfBound(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	chr.AddSubround(initSubroundHandlerMock())

	chr.SetSubroundHandlers(make([]chronology.SubroundHandler, 0))

	assert.Nil(t, chr.LoadSubroundHandler(0))
}

func TestChronology_HaveTimeInCurrentRoundShouldReturnPositiveValue(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	timeElapsed := int64(rounderMock.TimeDuration() / 10)

	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	haveTime := chr.HaveTimeInCurrentRound()

	assert.Equal(t, time.Duration(int64(rounderMock.TimeDuration())-timeElapsed), haveTime)
	assert.True(t, haveTime > 0)
}

func TestChronology_HaveTimeInCurrentRoundShouldReturnNegativeValue(t *testing.T) {
	rounderMock := &mock.RounderMock{}

	syncTimerMock := &mock.SyncTimerMock{}

	timeElapsed := int64(rounderMock.TimeDuration() * 2)

	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}

	genesisTime := time.Now()

	chr, _ := chronology.NewChronology(
		genesisTime,
		rounderMock,
		syncTimerMock)

	haveTime := chr.HaveTimeInCurrentRound()

	assert.Equal(t, time.Duration(int64(rounderMock.TimeDuration())-timeElapsed), haveTime)
	assert.True(t, haveTime < 0)
}
