package components

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func getMockArgManulRoundHandler() ArgManualRoundHandler {
	return ArgManualRoundHandler{
		EnableRoundsHandler:       &testscommon.EnableRoundsHandlerStub{},
		GenesisTimeStamp:          100,
		SupernovaGenesisTimeStamp: 120,
		RoundDuration:             time.Second,
		SupernovaRoundDuration:    time.Millisecond * 500,
		InitialRound:              0,
	}
}

func TestNewManualRoundHandler(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := NewManualRoundHandler(getMockArgManulRoundHandler())
		require.NotNil(t, handler)
		require.NoError(t, err)
	})
	t.Run("nil EnableRoundsHandler", func(t *testing.T) {
		t.Parallel()

		args := getMockArgManulRoundHandler()
		args.EnableRoundsHandler = nil
		handler, err := NewManualRoundHandler(args)
		require.Nil(t, handler)
		require.Equal(t, errNilEnableRoundsHandler, err)
	})
}

func TestManualRoundHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var handler *manualRoundHandler
	require.True(t, handler.IsInterfaceNil())

	handler, _ = NewManualRoundHandler(getMockArgManulRoundHandler())
	require.False(t, handler.IsInterfaceNil())
}

func TestManualRoundHandler_Operations(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()
	providedIndex := int64(0)
	providedRoundDuration := time.Second
	args := getMockArgManulRoundHandler()
	args.GenesisTimeStamp = genesisTime.UnixMilli()
	args.RoundDuration = providedRoundDuration
	args.InitialRound = providedIndex
	supernovaRound := uint64(20)
	args.SupernovaGenesisTimeStamp = genesisTime.Add(time.Duration(supernovaRound) * args.RoundDuration).UnixMilli()
	args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
		IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
			return round == supernovaRound
		},
		GetActivationRoundCalled: func(flag common.EnableRoundFlag) uint64 {
			return supernovaRound
		},
	}

	handler, _ := NewManualRoundHandler(args)
	require.Equal(t, providedIndex, handler.Index())
	handler.IncrementIndex()
	require.Equal(t, providedIndex+1, handler.Index())
	expectedTimestamp := time.UnixMilli(handler.genesisTimeStamp).Add(providedRoundDuration)
	require.Equal(t, expectedTimestamp, handler.TimeStamp())
	require.Equal(t, providedRoundDuration, handler.TimeDuration())
	providedMaxTime := time.Minute
	require.Equal(t, providedMaxTime, handler.RemainingTime(time.Now(), providedMaxTime))
	require.False(t, handler.BeforeGenesis())
	handler.UpdateRound(time.Now(), time.Now()) // for coverage only

	// revert the handler.IncrementIndex() call
	handler.RevertOneRound()

	// force supernova round
	for i := uint64(0); i < supernovaRound; i++ {
		handler.IncrementIndex()
	}
	require.Equal(t, args.SupernovaGenesisTimeStamp, handler.TimeStamp().UnixMilli())
}
