package v1_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapperStubs"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func defaultSubroundStartRoundFromSubround(sr *spos.Subround) (v1.SubroundStartRound, error) {
	startRound, err := v1.NewSubroundStartRound(
		sr,
		extend,
		v1.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
		&testscommon.SentSignatureTrackerStub{},
	)

	return startRound, err
}

func defaultWithoutErrorSubroundStartRoundFromSubround(sr *spos.Subround) v1.SubroundStartRound {
	startRound, _ := v1.NewSubroundStartRound(
		sr,
		extend,
		v1.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
		&testscommon.SentSignatureTrackerStub{},
	)

	return startRound
}

func defaultSubround(
	consensusState *spos.ConsensusState,
	ch chan bool,
	container spos.ConsensusCoreHandler,
) (*spos.Subround, error) {

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

func initSubroundStartRoundWithContainer(container spos.ConsensusCoreHandler) v1.SubroundStartRound {
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubround(consensusState, ch, container)
	srStartRound, _ := v1.NewSubroundStartRound(
		sr,
		extend,
		v1.ProcessingThresholdPercent,
		executeStoredMessages,
		resetConsensusMessages,
		&testscommon.SentSignatureTrackerStub{},
	)

	return srStartRound
}

func initSubroundStartRound() v1.SubroundStartRound {
	container := consensusMocks.InitConsensusCore()
	return initSubroundStartRoundWithContainer(container)
}

func TestNewSubroundStartRound(t *testing.T) {
	t.Parallel()

	ch := make(chan bool, 1)
	consensusState := initializers.InitConsensusState()
	container := consensusMocks.InitConsensusCore()
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

		srStartRound, err := v1.NewSubroundStartRound(
			nil,
			extend,
			v1.ProcessingThresholdPercent,
			executeStoredMessages,
			resetConsensusMessages,
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil extend function handler should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := v1.NewSubroundStartRound(
			sr,
			nil,
			v1.ProcessingThresholdPercent,
			executeStoredMessages,
			resetConsensusMessages,
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srStartRound)
		assert.ErrorIs(t, err, spos.ErrNilFunctionHandler)
		assert.Contains(t, err.Error(), "extend")
	})
	t.Run("nil executeStoredMessages function handler should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			nil,
			resetConsensusMessages,
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srStartRound)
		assert.ErrorIs(t, err, spos.ErrNilFunctionHandler)
		assert.Contains(t, err.Error(), "executeStoredMessages")
	})
	t.Run("nil resetConsensusMessages function handler should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			executeStoredMessages,
			nil,
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srStartRound)
		assert.ErrorIs(t, err, spos.ErrNilFunctionHandler)
		assert.Contains(t, err.Error(), "resetConsensusMessages")
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			executeStoredMessages,
			resetConsensusMessages,
			nil,
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, v1.ErrNilSentSignatureTracker, err)
	})
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetBlockchain(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetBootStrapper(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	sr.ConsensusStateHandler = nil
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetMultiSignerContainer(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetRoundHandler(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetSyncTimer(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundStartRound_NewSubroundStartRoundNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)
	container.SetNodesCoordinator(nil)
	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.Nil(t, srStartRound)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestSubroundStartRound_NewSubroundStartRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.NotNil(t, srStartRound)
	assert.Nil(t, err)
}

func TestSubroundStartRound_DoStartRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound := defaultWithoutErrorSubroundStartRoundFromSubround(sr)

	r := srStartRound.DoStartRoundJob()
	assert.True(t, r)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := initSubroundStartRound()

	sr.SetRoundCanceled(true)

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := initSubroundStartRound()

	sr.SetStatus(bls.SrStartRound, spos.SsFinished)

	ok := sr.DoStartRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundStartRound_DoStartRoundConsensusCheckShouldReturnTrueWhenInitCurrentRoundReturnTrue(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &bootstrapperStubs.BootstrapperStub{GetNodeStateCalled: func() common.NodeState {
		return common.NsSynchronized
	}}

	container := consensusMocks.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	sr := initSubroundStartRoundWithContainer(container)
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

	bootstrapperMock := &bootstrapperStubs.BootstrapperStub{GetNodeStateCalled: func() common.NodeState {
		return common.NsNotSynchronized
	}}

	container := consensusMocks.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)
	container.SetRoundHandler(initRoundHandlerMock())

	sr := initSubroundStartRoundWithContainer(container)

	ok := sr.DoStartRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetNodeStateNotReturnSynchronized(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &bootstrapperStubs.BootstrapperStub{}

	bootstrapperMock.GetNodeStateCalled = func() common.NodeState {
		return common.NsNotSynchronized
	}
	container := consensusMocks.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGenerateNextConsensusGroupErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}
	err := errors.New("error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte, round uint64, shardId uint32, epoch uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
		return nil, nil, err
	}
	container := consensusMocks.InitConsensusCore()
	container.SetNodesCoordinator(validatorGroupSelector)

	srStartRound := initSubroundStartRoundWithContainer(container)

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
	container := consensusMocks.InitConsensusCore()
	container.SetNodeRedundancyHandler(nodeRedundancyMock)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}
	leader := &shardingMocks.ValidatorMock{PubKeyCalled: func() []byte {
		return []byte("leader")
	}}

	validatorGroupSelector.ComputeValidatorsGroupCalled = func(
		bytes []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
		// will cause an error in GetLeader because of empty consensus group
		return leader, []nodesCoordinator.Validator{}, nil
	}

	container := consensusMocks.InitConsensusCore()
	container.SetNodesCoordinator(validatorGroupSelector)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrueWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	consensusState.SetSelfPubKey(consensusState.SelfPubKey() + "X")
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound := defaultWithoutErrorSubroundStartRoundFromSubround(sr)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenTimeIsOut(t *testing.T) {
	t.Parallel()

	roundHandlerMock := initRoundHandlerMock()

	roundHandlerMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Duration(-1)
	}

	container := consensusMocks.InitConsensusCore()
	container.SetRoundHandler(roundHandlerMock)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	bootstrapperMock := &bootstrapperStubs.BootstrapperStub{}

	bootstrapperMock.GetNodeStateCalled = func() common.NodeState {
		return common.NsSynchronized
	}

	container := consensusMocks.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.True(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldMetrics(t *testing.T) {
	t.Parallel()

	t.Run("not in consensus node", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		container := consensusMocks.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasCalled = true
					assert.Equal(t, "not in consensus group", value)
				}
			},
		}
		ch := make(chan bool, 1)
		consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
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

		srStartRound, _ := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			displayStatistics,
			executeStoredMessages,
			&testscommon.SentSignatureTrackerStub{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
	})
	t.Run("main key participant", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		wasIncrementCalled := false
		container := consensusMocks.InitConsensusCore()
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
		consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
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

		srStartRound, _ := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			displayStatistics,
			executeStoredMessages,
			&testscommon.SentSignatureTrackerStub{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
		assert.True(t, wasIncrementCalled)
	})
	t.Run("multi key participant", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		wasIncrementCalled := false
		container := consensusMocks.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
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
		consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
		consensusState.SetSelfPubKey("B")
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

		srStartRound, _ := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			displayStatistics,
			executeStoredMessages,
			&testscommon.SentSignatureTrackerStub{},
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
		container := consensusMocks.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasMetricConsensusStateCalled = true
					assert.Equal(t, "proposer", value)
				}
				if key == common.MetricConsensusRoundState {
					cntMetricConsensusRoundStateCalled++
					switch cntMetricConsensusRoundStateCalled {
					case 1:
						assert.Equal(t, "", value)
					case 2:
						assert.Equal(t, "proposed", value)
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
		consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
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

		srStartRound, _ := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			displayStatistics,
			executeStoredMessages,
			&testscommon.SentSignatureTrackerStub{},
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
		container := consensusMocks.InitConsensusCore()
		keysHandler := &testscommon.KeysHandlerStub{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				if key == common.MetricConsensusState {
					wasMetricConsensusStateCalled = true
					assert.Equal(t, "proposer", value)
				}
				if key == common.MetricConsensusRoundState {
					cntMetricConsensusRoundStateCalled++
					switch cntMetricConsensusRoundStateCalled {
					case 1:
						assert.Equal(t, "", value)
					case 2:
						assert.Equal(t, "proposed", value)
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
		consensusState := initializers.InitConsensusStateWithKeysHandler(keysHandler)
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

		srStartRound, _ := v1.NewSubroundStartRound(
			sr,
			extend,
			v1.ProcessingThresholdPercent,
			displayStatistics,
			executeStoredMessages,
			&testscommon.SentSignatureTrackerStub{},
		)
		srStartRound.Check()
		assert.True(t, wasMetricConsensusStateCalled)
		assert.True(t, wasMetricCountLeaderCalled)
		assert.Equal(t, 2, cntMetricConsensusRoundStateCalled)
	})
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldReturnErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}

	err := errors.New("error")
	validatorGroupSelector.ComputeValidatorsGroupCalled = func(
		bytes []byte,
		round uint64,
		shardId uint32,
		epoch uint32,
	) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
		return nil, nil, err
	}
	container := consensusMocks.InitConsensusCore()
	container.SetNodesCoordinator(validatorGroupSelector)

	srStartRound := initSubroundStartRoundWithContainer(container)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, err, err2)
}
