package v2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"

	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapperStubs"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
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

func defaultSubroundStartRoundFromSubround(sr *spos.Subround) (v2.SubroundStartRound, error) {
	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
	)

	return startRound, err
}

func defaultWithoutErrorSubroundStartRoundFromSubround(sr *spos.Subround) v2.SubroundStartRound {
	startRound, _ := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
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

func initSubroundStartRoundWithContainer(container spos.ConsensusCoreHandler) v2.SubroundStartRound {
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubround(consensusState, ch, container)
	srStartRound, _ := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
	)

	return srStartRound
}

func initSubroundStartRound() v2.SubroundStartRound {
	container := consensus.InitConsensusCore()
	return initSubroundStartRoundWithContainer(container)
}

func TestNewSubroundStartRound(t *testing.T) {
	t.Parallel()

	ch := make(chan bool, 1)
	consensusState := initializers.InitConsensusState()
	container := consensus.InitConsensusCore()
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

		srStartRound, err := v2.NewSubroundStartRound(
			nil,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			nil,
			&consensus.SposWorkerMock{},
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, v2.ErrNilSentSignatureTracker, err)
	})
	t.Run("nil worker should error", func(t *testing.T) {
		t.Parallel()

		srStartRound, err := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			nil,
		)

		assert.Nil(t, srStartRound)
		assert.Equal(t, spos.ErrNilWorker, err)
	})
}

func TestSubroundStartRound_NewSubroundStartRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()
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

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubround(consensusState, ch, container)

	srStartRound, err := defaultSubroundStartRoundFromSubround(sr)

	assert.NotNil(t, srStartRound)
	assert.Nil(t, err)
}

func TestSubroundStartRound_DoStartRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	container := consensus.InitConsensusCore()

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

	container := consensus.InitConsensusCore()
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

	container := consensus.InitConsensusCore()
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
	container := consensus.InitConsensusCore()
	container.SetBootStrapper(bootstrapperMock)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenGenerateNextConsensusGroupErr(t *testing.T) {
	t.Parallel()

	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}

	validatorGroupSelector.ComputeValidatorsGroupCalled = func(bytes []byte, round uint64, shardId uint32, epoch uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
		return nil, nil, expErr
	}
	container := consensus.InitConsensusCore()

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
	container := consensus.InitConsensusCore()
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

	container := consensus.InitConsensusCore()
	container.SetNodesCoordinator(validatorGroupSelector)

	srStartRound := initSubroundStartRoundWithContainer(container)

	r := srStartRound.InitCurrentRound()
	assert.False(t, r)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnTrueWhenIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	container := consensus.InitConsensusCore()
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

	container := consensus.InitConsensusCore()
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

	container := consensus.InitConsensusCore()
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
		container := consensus.InitConsensusCore()
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

		srStartRound, _ := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
	})
	t.Run("main key participant", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		wasIncrementCalled := false
		container := consensus.InitConsensusCore()
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

		srStartRound, _ := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasCalled)
		assert.True(t, wasIncrementCalled)
	})
	t.Run("multi key participant", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		wasIncrementCalled := false
		container := consensus.InitConsensusCore()
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

		srStartRound, _ := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
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
		container := consensus.InitConsensusCore()
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

		srStartRound, _ := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
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
		container := consensus.InitConsensusCore()
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

		srStartRound, _ := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
		)
		srStartRound.Check()
		assert.True(t, wasMetricConsensusStateCalled)
		assert.True(t, wasMetricCountLeaderCalled)
		assert.Equal(t, 2, cntMetricConsensusRoundStateCalled)
	})
}

func buildDefaultSubround(container spos.ConsensusCoreHandler) *spos.Subround {
	ch := make(chan bool, 1)
	consensusState := initializers.InitConsensusState()
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

	return sr
}

func TestSubroundStartRound_GenerateNextConsensusGroupShouldErrNilHeader(t *testing.T) {
	t.Parallel()

	container := consensus.InitConsensusCore()

	chainHandlerMock := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return nil
		},
	}

	container.SetBlockchain(chainHandlerMock)

	sr := buildDefaultSubround(container)
	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
	)
	require.Nil(t, err)

	err = startRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestSubroundStartRound_InitCurrentRoundShouldReturnFalseWhenResetErr(t *testing.T) {
	t.Parallel()

	container := consensus.InitConsensusCore()

	signingHandlerMock := &consensus.SigningHandlerStub{
		ResetCalled: func(pubKeys []string) error {
			return expErr
		},
	}

	container.SetSigningHandler(signingHandlerMock)

	sr := buildDefaultSubround(container)
	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
	)
	require.Nil(t, err)

	r := startRound.InitCurrentRound()

	assert.False(t, r)
}

func TestSubroundStartRound_IndexRoundIfNeededFailShardIdForEpoch(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := consensus.InitConsensusCore()

	idVar := 0

	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(idVar)
		},
	})

	container.SetNodesCoordinator(
		&shardingMocks.NodesCoordinatorStub{
			ShardIdForEpochCalled: func(epoch uint32) (uint32, error) {
				return 0, expErr
			},
		})

	sr := buildDefaultSubround(container)

	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
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

func TestSubroundStartRound_IndexRoundIfNeededGetValidatorsIndexesShouldNotBeCalled(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := consensus.InitConsensusCore()

	idVar := 0

	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(idVar)
		},
	})

	container.SetNodesCoordinator(
		&shardingMocks.NodesCoordinatorStub{
			GetValidatorsIndexesCalled: func(pubKeys []string, epoch uint32) ([]uint64, error) {
				require.Fail(t, "SaveRoundsInfo should not be called")
				return nil, expErr
			},
		})

	sr := buildDefaultSubround(container)

	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
	)
	require.Nil(t, err)

	called := false
	_ = startRound.SetOutportHandler(&outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			called = true
		},
	})

	startRound.IndexRoundIfNeeded(pubKeys)
	require.True(t, called)
}

func TestSubroundStartRound_IndexRoundIfNeededShouldFullyWork(t *testing.T) {

	pubKeys := []string{"testKey1", "testKey2"}

	container := consensus.InitConsensusCore()

	idVar := 0

	saveRoundInfoCalled := false

	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(idVar)
		},
	})

	sr := buildDefaultSubround(container)

	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
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

	container := consensus.InitConsensusCore()

	shardID := 1
	container.SetShardCoordinator(&processMock.CoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(shardID)
		},
	})

	container.SetNodesCoordinator(&shardingMocks.NodesCoordinatorStub{
		ShardIdForEpochCalled: func(epoch uint32) (uint32, error) {
			return 0, nil
		},
	})

	sr := buildDefaultSubround(container)

	startRound, err := v2.NewSubroundStartRound(
		sr,
		v2.ProcessingThresholdPercent,
		&testscommon.SentSignatureTrackerStub{},
		&consensus.SposWorkerMock{},
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

		container := consensus.InitConsensusCore()
		exErr := fmt.Errorf("expected error")
		container.SetNodesCoordinator(
			&shardingMocks.NodesCoordinatorStub{
				GetConsensusWhitelistedNodesCalled: func(epoch uint32) (map[string]struct{}, error) {
					return nil, exErr
				},
			})

		sr := buildDefaultSubround(container)

		startRound, err := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
		)
		require.Nil(t, err)
		startRound.ChangeEpoch(1)
	})
	t.Run("success - no panic", func(t *testing.T) {
		t.Parallel()

		defer expectNoPanic()

		container := consensus.InitConsensusCore()
		expectedKeys := map[string]struct{}{
			"aaa": {},
			"bbb": {},
		}

		container.SetNodesCoordinator(
			&shardingMocks.NodesCoordinatorStub{
				GetConsensusWhitelistedNodesCalled: func(epoch uint32) (map[string]struct{}, error) {
					return expectedKeys, nil
				},
			})

		sr := buildDefaultSubround(container)

		startRound, err := v2.NewSubroundStartRound(
			sr,
			v2.ProcessingThresholdPercent,
			&testscommon.SentSignatureTrackerStub{},
			&consensus.SposWorkerMock{},
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
	) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
		return nil, nil, expErr
	}
	container := consensus.InitConsensusCore()
	container.SetNodesCoordinator(validatorGroupSelector)

	srStartRound := initSubroundStartRoundWithContainer(container)

	err2 := srStartRound.GenerateNextConsensusGroup(0)

	assert.Equal(t, expErr, err2)
}
