package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	mock "github.com/multiversx/mx-chain-go/testscommon/epochstartmock"
	outportStub "github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func getDefaultArgumentsSubroundHandler() *SubroundsHandlerArgs {
	handlerArgs := &SubroundsHandlerArgs{
		Chronology:           &consensus.ChronologyHandlerMock{},
		ConsensusState:       &consensus.ConsensusStateMock{},
		Worker:               &consensus.SposWorkerMock{},
		SignatureThrottler:   &common.ThrottlerStub{},
		AppStatusHandler:     &statusHandler.AppStatusHandlerStub{},
		OutportHandler:       &outportStub.OutportStub{},
		SentSignatureTracker: &testscommon.SentSignatureTrackerStub{},
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ChainID:              []byte("chainID"),
		CurrentPid:           "peerID",
	}

	consensusCore := &consensus.ConsensusCoreMock{}
	consensusCore.SetEpochStartNotifier(&mock.EpochStartNotifierStub{})
	handlerArgs.ConsensusCoreHandler = consensusCore

	return handlerArgs
}

func TestNewSubroundsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil chronology should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.Chronology = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilChronologyHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil consensus core should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.ConsensusCoreHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilConsensusCoreHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil consensus state should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.ConsensusState = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilConsensusState, err)
		require.Nil(t, sh)
	})
	t.Run("nil worker should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.Worker = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilWorker, err)
		require.Nil(t, sh)
	})
	t.Run("nil signature throttler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.SignatureThrottler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilSignatureThrottler, err)
		require.Nil(t, sh)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.AppStatusHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilAppStatusHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil outport handler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.OutportHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilOutportHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil sent signature tracker should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.SentSignatureTracker = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilSentSignatureTracker, err)
		require.Nil(t, sh)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.EnableEpochsHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilEnableEpochsHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil chain ID should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.ChainID = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilChainID, err)
		require.Nil(t, sh)
	})
	t.Run("empty current PID should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		handlerArgs.CurrentPid = ""
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilCurrentPid, err)
		require.Nil(t, sh)
	})
	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		handlerArgs := getDefaultArgumentsSubroundHandler()
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
	})
}
