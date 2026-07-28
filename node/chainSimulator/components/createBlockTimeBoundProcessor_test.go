package components

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/round"
)

func TestCreateBlockTimeBoundProcessor_AppliesLimitToBothCreationPaths(t *testing.T) {
	t.Parallel()

	currentTime := time.Unix(100, 0)
	roundHandler := &round.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return 600 * time.Millisecond
		},
	}

	assertLimit := func(_ data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		require.True(t, haveTime())

		currentTime = currentTime.Add(149 * time.Millisecond)
		require.True(t, haveTime())

		currentTime = currentTime.Add(time.Millisecond)
		require.False(t, haveTime())

		return nil, nil, nil
	}

	blockProcessor := &testscommon.BlockProcessorStub{
		CreateBlockCalled:         assertLimit,
		CreateBlockProposalCalled: assertLimit,
	}
	processor := newCreateBlockTimeBoundProcessor(blockProcessor, roundHandler, 0)
	processor.now = func() time.Time {
		return currentTime
	}

	_, _, err := processor.CreateBlock(nil, func() bool { return true })
	require.NoError(t, err)

	currentTime = time.Unix(200, 0)
	_, _, err = processor.CreateBlockProposal(nil, func() bool { return true })
	require.NoError(t, err)
}

func TestCreateBlockTimeBoundProcessor_PreservesConsensusDeadline(t *testing.T) {
	t.Parallel()

	roundHandler := &round.RoundHandlerMock{
		TimeDurationCalled: func() time.Duration {
			return time.Second
		},
	}
	blockProcessor := &testscommon.BlockProcessorStub{
		CreateBlockProposalCalled: func(
			_ data.HeaderHandler,
			haveTime func() bool,
		) (data.HeaderHandler, data.BodyHandler, error) {
			require.False(t, haveTime())
			return nil, nil, nil
		},
	}

	processor := newCreateBlockTimeBoundProcessor(blockProcessor, roundHandler, 0.5)
	_, _, err := processor.CreateBlockProposal(nil, func() bool { return false })
	require.NoError(t, err)
}
