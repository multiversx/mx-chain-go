package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseSlashingDetector_IsRoundRelevant_DifferentRelevantAndIrrelevantRounds(t *testing.T) {
	t.Parallel()

	round := uint64(100)
	bsd := baseSlashingDetector{roundHandler: &mock.RoundHandlerMock{
		RoundIndex: int64(round),
	}}

	tests := []struct {
		round    func() uint64
		relevant bool
	}{
		{
			round: func() uint64 {
				return round
			},
			relevant: true,
		},
		{
			round: func() uint64 {
				return round - MaxDeltaToCurrentRound + 1
			},
			relevant: true,
		},
		{
			round: func() uint64 {
				return round + MaxDeltaToCurrentRound - 1
			},
			relevant: true,
		},
		{
			round: func() uint64 {
				return round - MaxDeltaToCurrentRound
			},
			relevant: false,
		},
		{
			round: func() uint64 {
				return round + MaxDeltaToCurrentRound
			},
			relevant: false,
		},
		{
			round: func() uint64 {
				return round - MaxDeltaToCurrentRound - 1
			},
			relevant: false,
		},
		{
			round: func() uint64 {
				return round + MaxDeltaToCurrentRound + 1
			},
			relevant: false,
		},
	}

	for _, currTest := range tests {
		require.Equal(t, currTest.relevant, bsd.isRoundRelevant(currTest.round()))
	}
}
