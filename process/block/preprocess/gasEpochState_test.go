package preprocess

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
)

func TestNewGasEpochState(t *testing.T) {
	t.Parallel()

	t.Run("nil economics fee handler", func(t *testing.T) {
		t.Parallel()

		ges, err := newGasEpochState(nil, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
		assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
		assert.True(t, check.IfNil(ges))
	})
	t.Run("nil enable epochs handler", func(t *testing.T) {
		t.Parallel()

		ges, err := newGasEpochState(feeHandlerMock(), nil, &testscommon.EnableRoundsHandlerStub{})
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
		assert.True(t, check.IfNil(ges))
	})
	t.Run("nil enable rounds handler", func(t *testing.T) {
		t.Parallel()

		ges, err := newGasEpochState(feeHandlerMock(), &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, nil)
		assert.Equal(t, process.ErrNilEnableRoundsHandler, err)
		assert.True(t, check.IfNil(ges))
	})
	t.Run("should create gas epoch state", func(t *testing.T) {
		t.Parallel()

		ges, err := newGasEpochState(feeHandlerMock(), &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ges))
	})
}

func TestGasEpochState_EpochConfirmed(t *testing.T) {
	t.Parallel()

	t.Run("should set epoch for limits", func(t *testing.T) {
		t.Parallel()

		ges, err := newGasEpochState(
			feeHandlerMock(),
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return true
				},
			},
			&testscommon.EnableRoundsHandlerStub{
				IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
					return true
				},
			})
		assert.Nil(t, err)

		assert.Equal(t, uint32(0), ges.epochForLimits)
		ges.EpochConfirmed(10)
		assert.Equal(t, uint32(10), ges.epochForLimits)
	})
	t.Run("should set previous epoch for limits if round flag not enabled", func(t *testing.T) {
		t.Parallel()

		ges, err := newGasEpochState(
			feeHandlerMock(),
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return true
				},
			},
			&testscommon.EnableRoundsHandlerStub{
				IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
					return false
				},
			})
		assert.Nil(t, err)

		assert.Equal(t, uint32(0), ges.epochForLimits)
		ges.EpochConfirmed(10)
		assert.Equal(t, uint32(9), ges.epochForLimits)
	})
}

func TestGasEpochState_RoundConfirmed(t *testing.T) {
	t.Parallel()

	feeHandler := feeHandlerMock()
	feeHandler.BlockCapacityOverestimationFactorCalled = func() uint64 {
		return 3
	}
	ges, err := newGasEpochState(
		feeHandler,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
				return true
			},
		})
	assert.Nil(t, err)

	assert.Equal(t, uint64(0), ges.roundForLimits)
	ges.RoundConfirmed(20)
	assert.Equal(t, uint64(20), ges.roundForLimits)
	assert.Equal(t, uint32(1), ges.epochForLimits)
	assert.Equal(t, uint64(3), ges.overEstimationFactor)

	epoch, overEstimatorFactor := ges.GetEpochForLimitsAndOverEstimationFactor()
	assert.Equal(t, uint32(1), epoch)
	assert.Equal(t, uint64(3), overEstimatorFactor)
}
