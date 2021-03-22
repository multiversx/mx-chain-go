package epochproviders

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func getUnixHandler(unix int64) func() int64 {
	return func() int64 {
		return unix
	}
}

func TestNewArithmeticEpochProvider_InvalidRoundsPerEpoch(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          0,
		RoundTimeInMilliseconds: 1,
		StartTime:               1,
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.True(t, errors.Is(err, ErrInvalidRoundsPerEpoch))
	assert.True(t, check.IfNil(aep))
}

func TestNewArithmeticEpochProvider_InvalidRoundTimeInMilliseconds(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          1,
		RoundTimeInMilliseconds: 0,
		StartTime:               1,
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.True(t, errors.Is(err, ErrInvalidRoundTimeInMilliseconds))
	assert.True(t, check.IfNil(aep))
}

func TestNewArithmeticEpochProvider_InvalidStartTime(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          1,
		RoundTimeInMilliseconds: 1,
		StartTime:               -1,
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.True(t, errors.Is(err, ErrInvalidStartTime))
	assert.True(t, check.IfNil(aep))
}

func TestNewArithmeticEpochProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          2400,
		RoundTimeInMilliseconds: 6000,
		StartTime:               time.Now().Unix(),
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(aep))
	assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())
}

func TestArithmeticEpochProvider_ComputeEpochAtGenesis(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          2400,
		RoundTimeInMilliseconds: 6000,
		StartTime:               1000,
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(0))
	assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(1000))
	assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15400))
	assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15401))
	assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15405))
	assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15406))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15412))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(29800))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(29806))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(29812))
	assert.Equal(t, uint32(2), aep.CurrentComputedEpoch())
}

func TestArithmeticEpochProvider_EpochStartActionNilHeader(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          2400,
		RoundTimeInMilliseconds: 6000,
		StartTime:               1000,
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(15500))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep.EpochStartAction(nil)

	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())
}

func TestArithmeticEpochProvider_EpochStartAction(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          2400,
		RoundTimeInMilliseconds: 6000,
		StartTime:               1000,
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(15500))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep.SetUnixHandler(getUnixHandler(17500))

	aep.EpochStartAction(&block.Header{
		Epoch:     1,
		TimeStamp: 3000,
	})

	assert.Equal(t, uint32(2), aep.CurrentComputedEpoch())
}

func TestArithmeticEpochProvider_EpochIsActiveInNetwork(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          1,
		RoundTimeInMilliseconds: 1,
		StartTime:               1,
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(1))

	aep.SetCurrentComputedEpoch(0)
	assert.True(t, aep.EpochIsActiveInNetwork(0))
	assert.True(t, aep.EpochIsActiveInNetwork(1))
	assert.True(t, aep.EpochIsActiveInNetwork(2))

	aep.SetCurrentComputedEpoch(1)
	assert.True(t, aep.EpochIsActiveInNetwork(0))
	assert.True(t, aep.EpochIsActiveInNetwork(1))
	assert.True(t, aep.EpochIsActiveInNetwork(2))

	aep.SetCurrentComputedEpoch(2)
	assert.False(t, aep.EpochIsActiveInNetwork(0))
	assert.True(t, aep.EpochIsActiveInNetwork(1))
	assert.True(t, aep.EpochIsActiveInNetwork(2))
	assert.True(t, aep.EpochIsActiveInNetwork(3))
}

func TestArithmeticEpochProvider_EpochStartPrepareShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic, %v", r))
		}
	}()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          1,
		RoundTimeInMilliseconds: 1,
		StartTime:               1,
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(1))

	aep.EpochStartPrepare(nil, nil)
}

func TestArithmeticEpochProvider_NotifyOrder(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		RoundsPerEpoch:          1,
		RoundTimeInMilliseconds: 1,
		StartTime:               1,
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(1))

	assert.Equal(t, uint32(core.CurrentNetworkEpochProvider), aep.NotifyOrder())
}

func TestArithmeticEpochProvider_EpochsDiff(t *testing.T) {
	t.Parallel()

	roundTimeInMilliseconds := uint64(5000)
	roundsPerEpoch := 500
	millisInASec := 1000

	currentTimeStamp := uint64(1600000000)

	epochDuration := uint64(roundsPerEpoch+1) * (roundTimeInMilliseconds / uint64(millisInASec))

	headerTimestampForNewEpoch := currentTimeStamp - (epochDuration)

	diffTimeStampInSeconds := currentTimeStamp - headerTimestampForNewEpoch
	diffTimeStampInMilliseconds := diffTimeStampInSeconds * 1000
	diffRounds := diffTimeStampInMilliseconds / roundTimeInMilliseconds
	diffEpochs := diffRounds / uint64(roundsPerEpoch+1)

	assert.Equal(t, uint64(1), diffEpochs)
}
