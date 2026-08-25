package process_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func TestResolveGasProcessingPolicy(t *testing.T) {
	t.Parallel()

	const (
		supernovaEpoch = uint32(2)
		supernovaRound = uint64(260)
		legacyGasLimit = uint64(1_500_000_000)
	)

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.SupernovaFlag && epoch >= supernovaEpoch
		},
	}
	enableRoundsHandler := &testscommon.EnableRoundsHandlerStub{
		IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
			return flag == common.SupernovaRoundFlag && round >= supernovaRound
		},
	}
	requestedEpoch := uint32(0)
	feeHandler := &economicsmocks.EconomicsHandlerMock{
		MaxGasLimitPerBlockInEpochCalled: func(_ uint32, epoch uint32) uint64 {
			requestedEpoch = epoch
			return legacyGasLimit
		},
	}

	tests := []struct {
		name        string
		epoch       uint32
		round       uint64
		isHeaderV3  bool
		hasOverride bool
	}{
		{name: "before drain", epoch: supernovaEpoch - 1, round: supernovaRound - 1},
		{name: "drain first round", epoch: supernovaEpoch, round: 0, hasOverride: true},
		{name: "drain last round", epoch: supernovaEpoch, round: supernovaRound - 1, hasOverride: true},
		{name: "activation round", epoch: supernovaEpoch, round: supernovaRound},
		{name: "V3 with drain metadata", epoch: supernovaEpoch, round: supernovaRound - 1, isHeaderV3: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestedEpoch = 0
			header := &testscommon.HeaderHandlerStub{
				EpochField: test.epoch,
				RoundField: test.round,
				IsHeaderV3Called: func() bool {
					return test.isHeaderV3
				},
			}

			policy, err := process.ResolveGasProcessingPolicy(
				header,
				enableEpochsHandler,
				enableRoundsHandler,
				feeHandler,
				0,
			)
			require.NoError(t, err)
			require.Equal(t, test.hasOverride, policy.HasMaxGasLimitPerBlock())
			if test.hasOverride {
				require.Equal(t, legacyGasLimit, policy.MaxGasLimitPerBlock())
				require.Equal(t, supernovaEpoch-1, requestedEpoch)
				return
			}

			require.Zero(t, requestedEpoch)
		})
	}
}

func TestResolveGasProcessingPolicyEpochZeroDrain(t *testing.T) {
	t.Parallel()

	header := &testscommon.HeaderHandlerStub{}
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return flag == common.SupernovaFlag
		},
	}
	enableRoundsHandler := &testscommon.EnableRoundsHandlerStub{}
	feeHandler := &economicsmocks.EconomicsHandlerMock{}

	_, err := process.ResolveGasProcessingPolicy(header, enableEpochsHandler, enableRoundsHandler, feeHandler, 0)
	require.ErrorIs(t, err, process.ErrInvalidValue)
}
