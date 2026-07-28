package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestBoundedWaitRoundHandler_RemainingTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		remaining time.Duration
		expected  time.Duration
	}{
		{name: "negative remains negative", remaining: -time.Millisecond, expected: -time.Millisecond},
		{name: "short duration is unchanged", remaining: 50 * time.Millisecond, expected: 50 * time.Millisecond},
		{name: "production duration is bounded", remaining: 6 * time.Second, expected: simulatedConsensusMaxWait},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			roundHandler := &boundedWaitRoundHandler{
				RoundHandler: &testscommon.RoundHandlerMock{
					RemainingTimeCalled: func(_ time.Time, _ time.Duration) time.Duration {
						return test.remaining
					},
				},
			}

			require.Equal(t, test.expected, roundHandler.RemainingTime(time.Time{}, time.Hour))
		})
	}
}

func TestBoundedWaitRoundHandler_ShouldReceiveConsensusMessage(t *testing.T) {
	t.Parallel()

	handler := &boundedWaitRoundHandler{}
	require.True(t, handler.shouldReceiveConsensusMessage(), "group is not known before START_ROUND")

	handler.setConsensusParticipant(false)
	require.False(t, handler.shouldReceiveConsensusMessage())

	handler.setConsensusParticipant(true)
	require.True(t, handler.shouldReceiveConsensusMessage())

	handler.resetConsensusParticipation()
	require.True(t, handler.shouldReceiveConsensusMessage(), "a new round starts in the unknown state")
}
