package bls_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"

	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"

	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

var expErr = fmt.Errorf("expected error")

func defaultSubroundStartRoundFromSubround(sr *spos.Subround) (bls.SubroundStartRound, error) {
	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)

	return startRound, err
}

func defaultWithoutErrorSubroundStartRoundFromSubround(sr *spos.Subround) bls.SubroundStartRound {
	startRound, _ := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)

	return startRound
}

func defaultSubroundWithContainer(
	container spos.ConsensusCoreHandler,
) *spos.Subround {
	consensusState := initConsensusState()
	sr, _ := defaultSubround(consensusState, container)
	return sr
}

func defaultSubround(
	consensusState *spos.ConsensusState,
	container spos.ConsensusCoreHandler,
) (*spos.Subround, error) {
	ch := make(chan bool, 1)

	return spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
}

func initSubroundStartRoundWithContainer(container spos.ConsensusCoreHandler) bls.SubroundStartRound {
	consensusState := initConsensusState()
	sr, _ := defaultSubround(consensusState, container)
	srStartRound, _ := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)

	return srStartRound
}

func initSubroundStartRound() bls.SubroundStartRound {
	container := mock.InitConsensusCore()
	return initSubroundStartRoundWithContainer(container)
}

func TestNewSubroundStartRound(t *testing.T) {
	t.Parallel()

	ch := make(chan bool, 1)
	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	sr, _ := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	t.Run("nil subround should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := bls.NewSubroundStartRound(
			nil,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			nil,
			&mock.SposWorkerMock{},
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, bls.ErrNilSentSignatureTracker, err)
	})
	t.Run("nil worker should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			nil,
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, spos.ErrNilWorker, err)
	})
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)
	container.SetBlockchain(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)
	container.SetBootStrapper(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)

	sr.ConsensusState = nil
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)
	container.SetMultiSignerContainer(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)
	container.SetRoundHandler(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)
	container.SetSyncTimer(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)
	container.SetValidatorGroupSelector(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestSubroundStartRound_NewSubroundStartRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)

	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.NotNil(t, srStartRound)
	assert.Nil(t, err)
}

func TestSubroundStartRound_DoStartRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := defaultSubround(consensusState, container)

	srStartRound := *defaultWithoutErrorSubroundStartRoundFromSubround(sr)

	r := srStartRound.DoStartRoundJob()
	assert.True(t, r)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundStartRound()

	sr.RoundCanceled = true

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundStartRound()

	sr.SetStatus(bls.SrStartRound, spos.SsFinished)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenInitCurrentRoundReturnTrue(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperStub{GetNodeStateCalled: func() common.NodeState {
		return common.NsSynchronized
	}}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	sr := *initSubroundStartRoundWithContainer(container)
	sentTrackerInterface := sr.GetSentSignatureTracker()
	sentTracker := sentTrackerInterface.(*testscommon.SentSignatureTrackerStub)
	startRoundCalled := false
	sentTracker.StartRoundCalled = func() {
		startRoundCalled = true
	}

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
	assert.True(t, startRoundCalled)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenInitCurrentRoundReturnFalse(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperStub{GetNodeStateCalled: func() common.NodeState {
		return common.NsNotSynchronized
	}}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)
	container.SetRoundHandler(initRoundHandlerMock())

	sr := *initSubroundStartRoundWithContainer(container)

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetNodeStateNotReturnSynchronized(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperStub{}

	bootstrapperMock.GetNodeStateCalled = func() common.NodeState {
		return common.NsNotSynchronized
	}
	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGenerateNextConsensusGroupErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}

	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte, round uint64, shardId uint32, epoch uint32) ([]nodesCoordinator.Validator, error) {
		return nil, expErr
	}
	container := mock.InitConsensusCore()

	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrueWhenMainMachineIsActive(t *testing.T) {
	t.Parallel()

	nodeRedundancyMock := &mock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool {
			return true
		},
	}
	container := mock.InitConsensusCore()
	container.SetNodeRedundancyHandler(nodeRedundancyMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(
		bytes []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) ([]nodesCoordinator.Validator, error) {
		return make([]nodesCoordinator.Validator, 0), nil
	}

	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrueWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	consensusState.SetSelfPubKey(consensusState.SelfPubKey() + "X")

	sr, _ := defaultSubround(consensusState, container)

	srStartRound := *defaultWithoutErrorSubroundStartRoundFromSubround(sr)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenTimeIsOut(t *testing.T) {
	t.Parallel()

	roundHandlerMock := initRoundHandlerMock()

	roundHandlerMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Duration(-1)
	}

	container := mock.InitConsensusCore()
	container.SetRoundHandler(roundHandlerMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &mock.BootstrapperStub{}

	bootstrapperMock.GetNodeStateCalled = func() common.NodeState {
		return common.NsSynchronized
	}

	container := mock.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldMetrics(t *testing.T) {
	t.Parallel()

	t.Run("not in consensus node", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		container := mock.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasCalled = true
					assert.Equal(t, value, "not in consensus group")
				}
			},
		}
		ch := make(chan bool, 1)
		consensusState := initConsensusStateWithKeysHandler(keysHandler)
		consensusState.SetSelfPubKey("not in consensus")
		sr, _ := spos.NewSubround(
			-1,
			bls.SrStartRound,
			bls.SrBlock,
			int64(85*roundTimeDuration/100),
			int64(95*roundTimeDuration/100),
			"(START_ROUND)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			appStatusHandler,
		)

		srStartRound, _ := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
	})
	t.Run("main key participant", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		wasIncrementCalled := false
		container := mock.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return string(pkBytes) == "B"
			},
		}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasCalled = true
					assert.Equal(t, "participant", value)
				}
			},
			IncrementHandler: func(key string) {
				if key == common.MetricCountConsensus {
					wasIncrementCalled = true
				}
			},
		}
		ch := make(chan bool, 1)
		consensusState := initConsensusStateWithKeysHandler(keysHandler)
		consensusState.SetSelfPubKey("B")
		sr, _ := spos.NewSubround(
			-1,
			bls.SrStartRound,
			bls.SrBlock,
			int64(85*roundTimeDuration/100),
			int64(95*roundTimeDuration/100),
			"(START_ROUND)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			appStatusHandler,
		)

		srStartRound, _ := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
		assert.True(t, wasIncrementCalled)
	})
	t.Run("multi key participant", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		wasIncrementCalled := false
		container := mock.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasCalled = true
					assert.Equal(t, value, "participant")
				}
			},
			IncrementHandler: func(key string) {
				if key == common.MetricCountConsensus {
					wasIncrementCalled = true
				}
			},
		}
		ch := make(chan bool, 1)
		consensusState := initConsensusStateWithKeysHandler(keysHandler)
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return string(pkBytes) == consensusState.SelfPubKey()
		}
		sr, _ := spos.NewSubround(
			-1,
			bls.SrStartRound,
			bls.SrBlock,
			int64(85*roundTimeDuration/100),
			int64(95*roundTimeDuration/100),
			"(START_ROUND)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			appStatusHandler,
		)

		srStartRound, _ := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
		assert.True(t, wasIncrementCalled)
	})
	t.Run("main key leader", func(t *testing.T) {
		t.Parallel()

		wasMetricConsensusStateCalled := false
		wasMetricCountLeaderCalled := false
		cntMetricConsensusRoundStateCalled := 0
		container := mock.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasMetricConsensusStateCalled = true
					assert.Equal(t, value, "proposer")
				}
				if key == common.MetricConsensusRoundState {
					cntMetricConsensusRoundStateCalled++
					switch cntMetricConsensusRoundStateCalled {
					case 1:
						assert.Equal(t, value, "")
					case 2:
						assert.Equal(t, value, "proposed")
					default:
						assert.Fail(t, "should have been called only twice")
					}
				}
			},
			IncrementHandler: func(key string) {
				if key == common.MetricCountLeader {
					wasMetricCountLeaderCalled = true
				}
			},
		}
		ch := make(chan bool, 1)
		consensusState := initConsensusStateWithKeysHandler(keysHandler)
		leader, _ := consensusState.GetLeader()
		consensusState.SetSelfPubKey(leader)
		sr, _ := spos.NewSubround(
			-1,
			bls.SrStartRound,
			bls.SrBlock,
			int64(85*roundTimeDuration/100),
			int64(95*roundTimeDuration/100),
			"(START_ROUND)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			appStatusHandler,
		)

		srStartRound, _ := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasMetricConsensusStateCalled)
		assert.True(t, wasMetricCountLeaderCalled)
		assert.Equal(t, 2, cntMetricConsensusRoundStateCalled)
	})
	t.Run("managed key leader", func(t *testing.T) {
		t.Parallel()

		wasMetricConsensusStateCalled := false
		wasMetricCountLeaderCalled := false
		cntMetricConsensusRoundStateCalled := 0
		container := mock.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasMetricConsensusStateCalled = true
					assert.Equal(t, value, "proposer")
				}
				if key == common.MetricConsensusRoundState {
					cntMetricConsensusRoundStateCalled++
					switch cntMetricConsensusRoundStateCalled {
					case 1:
						assert.Equal(t, value, "")
					case 2:
						assert.Equal(t, value, "proposed")
					default:
						assert.Fail(t, "should have been called only twice")
					}
				}
			},
			IncrementHandler: func(key string) {
				if key == common.MetricCountLeader {
					wasMetricCountLeaderCalled = true
				}
			},
		}
		ch := make(chan bool, 1)
		consensusState := initConsensusStateWithKeysHandler(keysHandler)
		leader, _ := consensusState.GetLeader()
		consensusState.SetSelfPubKey(leader)
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return string(pkBytes) == leader
		}
		sr, _ := spos.NewSubround(
			-1,
			bls.SrStartRound,
			bls.SrBlock,
			int64(85*roundTimeDuration/100),
			int64(95*roundTimeDuration/100),
			"(START_ROUND)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			appStatusHandler,
		)

		srStartRound, _ := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasMetricConsensusStateCalled)
		assert.True(t, wasMetricCountLeaderCalled)
		assert.Equal(t, 2, cntMetricConsensusRoundStateCalled)
	})
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldErrNilHeader(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	chainHandlerMock := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return nil
		},
	}

	container.SetBlockchain(chainHandlerMock)

	sr := defaultSubroundWithContainer(container)
	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)
	require.Nil(t, err)

	err = startRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenResetErr(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	signingHandlerMock := &consensus.SigningHandlerStub{
		ResetCalled: func(pubKeys []string) error {
			return expErr
		},
	}

	container.SetSigningHandler(signingHandlerMock)

	sr := defaultSubroundWithContainer(container)
	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)
	require.Nil(t, err)

	r := startRound.InitCurrentRound()

	assert.False(t, r)
}

func TestSubroundStartRound_IndexRoundIfNeededFailShardIdForEpoch(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := mock.InitConsensusCore()

	idVar := 0

	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(idVar)
		},
	})

	container.SetValidatorGroupSelector(
		&shardingMocks.NodesCoordinatorStub{
			ShardIdForEpochCalled: func(epoch uint32) (uint32, error) {
				return 0, expErr
			},
		})

	sr := defaultSubroundWithContainer(container)

	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)
	require.Nil(t, err)

	_ = startRound.SetOutportHandler(&outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			require.Fail(t, "SaveRoundsInfo should not be called")
		},
	})

	startRound.IndexRoundIfNeeded(pubKeys)

}

func TestSubroundStartRound_IndexRoundIfNeededFailGetValidatorsIndexes(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := mock.InitConsensusCore()

	idVar := 0

	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(idVar)
		},
	})

	container.SetValidatorGroupSelector(
		&shardingMocks.NodesCoordinatorStub{
			GetValidatorsIndexesCalled: func(pubKeys []string, epoch uint32) ([]uint64, error) {
				return nil, expErr
			},
		})

	sr := defaultSubroundWithContainer(container)

	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)
	require.Nil(t, err)

	_ = startRound.SetOutportHandler(&outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			require.Fail(t, "SaveRoundsInfo should not be called")
		},
	})

	startRound.IndexRoundIfNeeded(pubKeys)

}

func TestSubroundStartRound_IndexRoundIfNeededShouldFullyWork(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := mock.InitConsensusCore()

	idVar := 0

	saveRoundInfoCalled := false

	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(idVar)
		},
	})

	sr := defaultSubroundWithContainer(container)

	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)
	require.Nil(t, err)

	_ = startRound.SetOutportHandler(&outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			saveRoundInfoCalled = true
		}})

	startRound.IndexRoundIfNeeded(pubKeys)

	assert.True(t, saveRoundInfoCalled)

}

func TestSubroundStartRound_IndexRoundIfNeededDifferentShardIdFail(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := mock.InitConsensusCore()

	shardID := 1
	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(shardID)
		},
	})

	container.SetValidatorGroupSelector(&shardingMocks.NodesCoordinatorStub{
		ShardIdForEpochCalled: func(epoch uint32) (uint32, error) {
			return 0, nil
		},
	})

	sr := defaultSubroundWithContainer(container)

	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{},
	)
	require.Nil(t, err)

	_ = startRound.SetOutportHandler(&outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			require.Fail(t, "SaveRoundsInfo should not be called")
		},
	})

	startRound.IndexRoundIfNeeded(pubKeys)

}

func TestSubroundStartRound_changeEpoch(t *testing.T) {
	t.Parallel()

	expectPanic := func() {
		if recover() == nil {
			require.Fail(t, "expected panic")
		}
	}

	expectNoPanic := func() {
		if recover() != nil {
			require.Fail(t, "expected no panic")
		}
	}

	t.Run("error returned by nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		defer expectPanic()

		container := mock.InitConsensusCore()
		exErr := fmt.Errorf("expected error")
		container.SetValidatorGroupSelector(
			&shardingMocks.NodesCoordinatorStub{
				GetConsensusWhitelistedNodesCalled: func(epoch uint32) (map[string]struct{}, error) {
					return nil, exErr
				},
			})

		sr := defaultSubroundWithContainer(container)

		startRound, err := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		require.Nil(t, err)
		startRound.ChangeEpoch(1)
	})
	t.Run("success - no panic", func(t *testing.T) {
		t.Parallel()

		defer expectNoPanic()

		container := mock.InitConsensusCore()
		expectedKeys := map[string]struct{}{
			"aaa": {},
			"bbb": {},
		}

		container.SetValidatorGroupSelector(
			&shardingMocks.NodesCoordinatorStub{
				GetConsensusWhitelistedNodesCalled: func(epoch uint32) (map[string]struct{}, error) {
					return expectedKeys, nil
				},
			})

		sr := defaultSubroundWithContainer(container)

		startRound, err := bls.NewSubroundStartRound(
			sr,
			bls.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&mock.SposWorkerMock{},
		)
		require.Nil(t, err)
		startRound.ChangeEpoch(1)
	})
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldReturnErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}

	validatorGroupSelector.ComputeValidatorsGroupCalled = func(
		bytes []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) ([]nodesCoordinator.Validator, error) {
		return nil, expErr
	}
	container := mock.InitConsensusCore()
	container.SetValidatorGroupSelector(validatorGroupSelector)

	srStartRound := *initSubroundStartRoundWithContainer(container)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, expErr, err2)
}

func TestSubroundStartRound_SetEquivalentProofForPreviousBlock(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	bitmap := []byte("bitmap")
	aggSig := []byte("aggSig")

	aggWasCalled := false
	signingHandlerMock := &consensus.SigningHandlerStub{
		AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
			aggWasCalled = true
			return aggSig, nil
		},
	}
	container.SetSigningHandler(signingHandlerMock)

	sr := defaultSubroundWithContainer(container)

	setProofCalled := false
	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{
			SetValidEquivalentProofCalled: func(headerHash []byte, proof data.HeaderProof) {
				require.Equal(t, data.HeaderProof{
					AggregatedSignature: aggSig,
					PubKeysBitmap:       bitmap,
				}, proof)

				setProofCalled = true
			},
		},
	)
	require.Nil(t, err)

	startRound.SetEquivalentProofForPreviousBlock(bitmap, 2)

	require.True(t, aggWasCalled)
	require.True(t, setProofCalled)
}

func TestSubroundStartRound_InitCurrentRound_ShouldSetPreviousEquivalentProof(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	container.SetEnableEpochsHandler(&enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.EquivalentMessagesFlag
		},
	})

	consensusState := initConsensusState()

	leader, _ := consensusState.GetLeader()
	consensusState.SetSelfPubKey(leader)

	sr, _ := defaultSubround(consensusState, container)

	setProofCalled := false
	startRound, err := bls.NewSubroundStartRound(
		sr,
		bls.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&mock.SposWorkerMock{
			SetValidEquivalentProofCalled: func(headerHash []byte, proof data.HeaderProof) {
				setProofCalled = true
			},
		},
	)
	require.Nil(t, err)

	r := startRound.InitCurrentRound()
	assert.True(t, r)

	require.True(t, setProofCalled)
}
