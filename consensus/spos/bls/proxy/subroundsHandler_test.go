package proxy

import (
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/require"

	mock2 "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapperStubs"
	"github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	mock "github.com/multiversx/mx-chain-go/testscommon/epochstartmock"
	outportStub "github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func getDefaultArgumentsSubroundHandler() (*SubroundsHandlerArgs, *spos.ConsensusCore) {
	x := make(chan bool)
	chronology := &consensus.ChronologyHandlerMock{}
	epochsEnable := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	consensusState := &consensus.ConsensusStateMock{}
	epochNotifier := &epochNotifierMock.EpochNotifierStub{}
	worker := &consensus.SposWorkerMock{
		RemoveAllReceivedMessagesCallsCalled: func() {},
		GetConsensusStateChangedChannelsCalled: func() chan bool {
			return x
		},
	}
	antiFloodHandler := &mock2.P2PAntifloodHandlerStub{}
	handlerArgs := &SubroundsHandlerArgs{
		Chronology:           chronology,
		ConsensusState:       consensusState,
		Worker:               worker,
		SignatureThrottler:   &common.ThrottlerStub{},
		AppStatusHandler:     &statusHandler.AppStatusHandlerStub{},
		OutportHandler:       &outportStub.OutportStub{},
		SentSignatureTracker: &testscommon.SentSignatureTrackerStub{},
		EnableEpochsHandler:  epochsEnable,
		ChainID:              []byte("chainID"),
		CurrentPid:           "peerID",
	}

	consensusCore := &spos.ConsensusCore{}
	consensusCore.SetEpochStartNotifier(epochStartNotifier)
	consensusCore.SetBlockchain(&testscommon.ChainHandlerStub{})
	consensusCore.SetBlockProcessor(&testscommon.BlockProcessorStub{})
	consensusCore.SetBootStrapper(&bootstrapperStubs.BootstrapperStub{})
	consensusCore.SetBroadcastMessenger(&consensus.BroadcastMessengerMock{})
	consensusCore.SetChronology(chronology)
	consensusCore.SetAntifloodHandler(antiFloodHandler)
	consensusCore.SetHasher(&testscommon.HasherStub{})
	consensusCore.SetMarshalizer(&testscommon.MarshallerStub{})
	consensusCore.SetMultiSignerContainer(&cryptoMocks.MultiSignerContainerStub{
		GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
			return &cryptoMocks.MultisignerMock{}, nil
		},
	})
	consensusCore.SetRoundHandler(&consensus.RoundHandlerMock{})
	consensusCore.SetShardCoordinator(&testscommon.ShardsCoordinatorMock{})
	consensusCore.SetSyncTimer(&testscommon.SyncTimerStub{})
	consensusCore.SetNodesCoordinator(&shardingMocks.NodesCoordinatorMock{})
	consensusCore.SetPeerHonestyHandler(&testscommon.PeerHonestyHandlerStub{})
	consensusCore.SetHeaderSigVerifier(&consensus.HeaderSigVerifierMock{})
	consensusCore.SetFallbackHeaderValidator(&testscommon.FallBackHeaderValidatorStub{})
	consensusCore.SetNodeRedundancyHandler(&mock2.NodeRedundancyHandlerStub{})
	consensusCore.SetScheduledProcessor(&consensus.ScheduledProcessorStub{})
	consensusCore.SetMessageSigningHandler(&mock2.MessageSigningHandlerStub{})
	consensusCore.SetPeerBlacklistHandler(&mock2.PeerBlacklistHandlerStub{})
	consensusCore.SetSigningHandler(&consensus.SigningHandlerStub{})
	consensusCore.SetEnableEpochsHandler(epochsEnable)
	consensusCore.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{})
	consensusCore.SetEpochNotifier(epochNotifier)
	consensusCore.SetInvalidSignersCache(&consensus.InvalidSignersCacheMock{})
	handlerArgs.ConsensusCoreHandler = consensusCore

	return handlerArgs, consensusCore
}

func TestNewSubroundsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil chronology should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.Chronology = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilChronologyHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil consensus core should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.ConsensusCoreHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilConsensusCoreHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil consensus state should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.ConsensusState = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilConsensusState, err)
		require.Nil(t, sh)
	})
	t.Run("nil worker should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.Worker = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilWorker, err)
		require.Nil(t, sh)
	})
	t.Run("nil signature throttler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.SignatureThrottler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilSignatureThrottler, err)
		require.Nil(t, sh)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.AppStatusHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilAppStatusHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil outport handler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.OutportHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilOutportHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil sent signature tracker should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.SentSignatureTracker = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilSentSignatureTracker, err)
		require.Nil(t, sh)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.EnableEpochsHandler = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilEnableEpochsHandler, err)
		require.Nil(t, sh)
	})
	t.Run("nil chain ID should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.ChainID = nil
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilChainID, err)
		require.Nil(t, sh)
	})
	t.Run("empty current PID should error", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		handlerArgs.CurrentPid = ""
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Equal(t, ErrNilCurrentPid, err)
		require.Nil(t, sh)
	})
	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
	})
}

func TestSubroundsHandler_initSubroundsForEpoch(t *testing.T) {
	t.Parallel()

	t.Run("equivalent messages not enabled, with previous consensus type not consensusV1", func(t *testing.T) {
		t.Parallel()

		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on register to EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())
		sh.currentConsensusType = consensusNone

		err = sh.initSubroundsForEpoch(0)
		require.Nil(t, err)
		require.Equal(t, consensusV1, sh.currentConsensusType)
		require.Equal(t, int32(2), startCalled.Load())
	})
	t.Run("equivalent messages not enabled, with previous consensus type consensusV1", func(t *testing.T) {
		t.Parallel()

		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on register to EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())
		sh.currentConsensusType = consensusV1

		err = sh.initSubroundsForEpoch(0)
		require.Nil(t, err)
		require.Equal(t, consensusV1, sh.currentConsensusType)
		require.Equal(t, int32(1), startCalled.Load())

	})
	t.Run("equivalent messages enabled, with previous consensus type consensusNone", func(t *testing.T) {
		t.Parallel()
		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on register to EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())
		sh.currentConsensusType = consensusNone

		err = sh.initSubroundsForEpoch(0)
		require.Nil(t, err)
		require.Equal(t, consensusV2, sh.currentConsensusType)
		require.Equal(t, int32(2), startCalled.Load())
	})
	t.Run("equivalent messages enabled, with previous consensus type consensusV1", func(t *testing.T) {
		t.Parallel()
		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on register to EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())
		sh.currentConsensusType = consensusV1

		err = sh.initSubroundsForEpoch(0)
		require.Nil(t, err)
		require.Equal(t, consensusV2, sh.currentConsensusType)
		require.Equal(t, int32(2), startCalled.Load())
	})
	t.Run("equivalent messages enabled, with previous consensus type consensusV2", func(t *testing.T) {
		t.Parallel()

		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on register to EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())
		sh.currentConsensusType = consensusV2

		err = sh.initSubroundsForEpoch(0)
		require.Nil(t, err)
		require.Equal(t, consensusV2, sh.currentConsensusType)
		require.Equal(t, int32(1), startCalled.Load())
	})
}

func TestSubroundsHandler_Start(t *testing.T) {
	t.Parallel()

	// the Start is tested via initSubroundsForEpoch, adding one of the test cases here as well
	t.Run("equivalent messages not enabled, with previous consensus type not consensusV1", func(t *testing.T) {
		t.Parallel()

		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on init of EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())
		sh.currentConsensusType = consensusNone

		err = sh.Start(0)
		require.Nil(t, err)
		require.Equal(t, consensusV1, sh.currentConsensusType)
		require.Equal(t, int32(2), startCalled.Load())
	})
}

func TestSubroundsHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	t.Run("nil handler", func(t *testing.T) {
		t.Parallel()

		var sh *SubroundsHandler
		require.True(t, sh.IsInterfaceNil())
	})
	t.Run("not nil handler", func(t *testing.T) {
		t.Parallel()

		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)

		require.False(t, sh.IsInterfaceNil())
	})
}

func TestSubroundsHandler_EpochConfirmed(t *testing.T) {
	t.Parallel()

	t.Run("nil handler does not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("The code panicked")
			}
		}()
		handlerArgs, _ := getDefaultArgumentsSubroundHandler()
		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		sh.EpochConfirmed(0, 0)
	})

	// tested through initSubroundsForEpoch
	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		startCalled := atomic.Int32{}
		handlerArgs, consensusCore := getDefaultArgumentsSubroundHandler()
		chronology := &consensus.ChronologyHandlerMock{
			StartRoundCalled: func() {
				startCalled.Add(1)
			},
		}
		enableEpoch := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
		}
		handlerArgs.Chronology = chronology
		handlerArgs.EnableEpochsHandler = enableEpoch
		consensusCore.SetEnableEpochsHandler(enableEpoch)
		consensusCore.SetChronology(chronology)

		sh, err := NewSubroundsHandler(handlerArgs)
		require.Nil(t, err)
		require.NotNil(t, sh)
		// first call on register to EpochNotifier
		require.Equal(t, int32(1), startCalled.Load())

		sh.currentConsensusType = consensusNone
		sh.EpochConfirmed(0, 0)
		require.Nil(t, err)
		require.Equal(t, consensusV1, sh.currentConsensusType)
		require.Equal(t, int32(2), startCalled.Load())
	})
}
