package epochproviders

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	commonErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
)

func getUnixHandler(unix int64) func() int64 {
	return func() int64 {
		return unix
	}
}

func getMockChainParametersHandler() *chainParameters.ChainParametersHandlerStub {
	return &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{
				RoundsPerEpoch: 2400,
				RoundDuration:  6000,
			}
		},
	}
}

func TestNewArithmeticEpochProvider_NilChainParameterHandler(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		ChainParametersHandler: nil,
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		StartTime:              1,
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.True(t, errors.Is(err, process.ErrNilChainParametersHandler))
	assert.True(t, check.IfNil(aep))
}

func TestNewArithmeticEpochProvider_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		ChainParametersHandler: getMockChainParametersHandler(),
		EnableEpochsHandler:    nil,
		StartTime:              1,
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.True(t, errors.Is(err, commonErrors.ErrNilEnableEpochsHandler))
	assert.True(t, check.IfNil(aep))
}

func TestNewArithmeticEpochProvider_InvalidStartTime(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		ChainParametersHandler: getMockChainParametersHandler(),
		StartTime:              -1,
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}

	aep, err := NewArithmeticEpochProvider(arg)

	assert.True(t, errors.Is(err, ErrInvalidStartTime))
	assert.True(t, check.IfNil(aep))
}

func TestNewArithmeticEpochProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		arg := ArgArithmeticEpochProvider{
			ChainParametersHandler: getMockChainParametersHandler(),
			StartTime:              time.Now().Unix(),
			EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}

		aep, err := NewArithmeticEpochProvider(arg)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(aep))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())
	})

	t.Run("after supernova", func(t *testing.T) {
		t.Parallel()

		arg := ArgArithmeticEpochProvider{
			ChainParametersHandler: getMockChainParametersHandler(),
			StartTime:              time.Now().UnixMilli(),
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.SupernovaFlag
				},
			},
		}

		aep, err := NewArithmeticEpochProvider(arg)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(aep))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())
	})
}

func TestArithmeticEpochProvider_ComputeEpochAtGenesis(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		arg := ArgArithmeticEpochProvider{
			ChainParametersHandler: getMockChainParametersHandler(),
			StartTime:              1000,
			EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
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
	})

	t.Run("after supernova", func(t *testing.T) {
		t.Parallel()

		arg := ArgArithmeticEpochProvider{
			ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
				CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
					return config.ChainParametersByEpochConfig{
						RoundsPerEpoch: 2400,
						RoundDuration:  6000,
					}
				},
			},
			StartTime: 1000 * 1000,
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.SupernovaFlag
				},
			},
		}
		aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(0))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(1000*1000))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15400*1000))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15401*1000))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15405*1000))
		assert.Equal(t, uint32(0), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15406*1000))
		assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(15412*1000))
		assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(29800*1000))
		assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(29806*1000))
		assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

		aep = NewTestArithmeticEpochProvider(arg, getUnixHandler(29812*1000))
		assert.Equal(t, uint32(2), aep.CurrentComputedEpoch())
	})
}

func TestArithmeticEpochProvider_EpochConfirmedInvalidTimestamp(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
			CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
				return config.ChainParametersByEpochConfig{
					RoundsPerEpoch: 2400,
					RoundDuration:  6000,
				}
			},
		},
		StartTime:           1000,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(15500))
	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

	aep.EpochConfirmed(1000, 0)

	assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())
}

func TestArithmeticEpochProvider_EpochConfirmed(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		arg := ArgArithmeticEpochProvider{
			ChainParametersHandler: getMockChainParametersHandler(),
			StartTime:              1000,
			EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		}
		aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(15500))
		assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

		aep.SetUnixHandler(getUnixHandler(17500))

		aep.EpochConfirmed(1, 3000)

		assert.Equal(t, uint32(2), aep.CurrentComputedEpoch())
	})

	t.Run("after supernova", func(t *testing.T) {
		t.Parallel()

		arg := ArgArithmeticEpochProvider{
			ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
				CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
					return config.ChainParametersByEpochConfig{
						RoundsPerEpoch: 2400,
						RoundDuration:  6000,
					}
				},
			},
			StartTime: 1000 * int64(millisecondsInOneSecond),
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.SupernovaFlag
				},
			},
		}
		aep := NewTestArithmeticEpochProvider(arg, getUnixHandler(15500*int64(millisecondsInOneSecond)))
		assert.Equal(t, uint32(1), aep.CurrentComputedEpoch())

		aep.SetUnixHandler(getUnixHandler(17500 * int64(millisecondsInOneSecond)))

		aep.EpochConfirmed(1, 3000*millisecondsInOneSecond)

		assert.Equal(t, uint32(2), aep.CurrentComputedEpoch())
	})
}

func TestArithmeticEpochProvider_EpochIsActiveInNetwork(t *testing.T) {
	t.Parallel()

	arg := ArgArithmeticEpochProvider{
		ChainParametersHandler: getMockChainParametersHandler(),
		StartTime:              1,
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
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
