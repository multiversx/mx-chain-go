package spos_test

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

var chainID = []byte("chain ID")
var wrongChainID = []byte("wrong chain ID")

const currentPid = core.PeerID("pid")

// executeStoredMessages tries to execute all the messages received which are valid for execution
func executeStoredMessages() {
}

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		var value string
		for j := 0; j < PublicKeySize; j++ {
			value += string([]byte{byte(i + 65)})
		}

		eligibleList = append(eligibleList, value)
	}

	return eligibleList
}

func initConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)

	eligibleNodesKeys := make(map[string]struct{}, len(eligibleList))
	for _, key := range eligibleList {
		eligibleNodesKeys[key] = struct{}{}
	}

	indexLeader := 1
	rcns, _ := spos.NewRoundConsensus(
		eligibleNodesKeys,
		consensusGroupSize,
		eligibleList[indexLeader],
		&testscommon.KeysHandlerStub{},
	)

	rcns.SetConsensusGroup(eligibleList)
	rcns.SetLeader(eligibleList[indexLeader])
	rcns.ResetRoundState()

	pBFTThreshold := consensusGroupSize*2/3 + 1
	pBFTFallbackThreshold := consensusGroupSize*1/2 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(1, 1)
	rthr.SetThreshold(2, pBFTThreshold)
	rthr.SetThreshold(3, pBFTThreshold)
	rthr.SetThreshold(4, pBFTThreshold)
	rthr.SetThreshold(5, pBFTThreshold)
	rthr.SetFallbackThreshold(1, 1)
	rthr.SetFallbackThreshold(2, pBFTFallbackThreshold)
	rthr.SetFallbackThreshold(3, pBFTFallbackThreshold)
	rthr.SetFallbackThreshold(4, pBFTFallbackThreshold)
	rthr.SetFallbackThreshold(5, pBFTFallbackThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")
	cns.SetRoundIndex(0)
	return cns
}

func TestSubround_NewSubroundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := consensus.InitConsensusCore()
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		nil,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	assert.Equal(t, spos.ErrNilConsensusState, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilChannelShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := consensus.InitConsensusCore()

	sr, err := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		nil,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	assert.Equal(t, spos.ErrNilChannel, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilExecuteStoredMessagesShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := consensus.InitConsensusCore()
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		nil,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	assert.Equal(t, spos.ErrNilExecuteStoredMessages, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundNilContainerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		int64(0*roundTimeDuration/100),
		int64(5*roundTimeDuration/100),
		"(START_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		nil,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	assert.Equal(t, spos.ErrNilConsensusCore, err)
	assert.Nil(t, sr)
}

func TestSubround_NilContainerBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetBlockchain(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubround_NilContainerBlockprocessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetBlockProcessor(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubround_NilContainerBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetBootStrapper(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestSubround_NilContainerChronologyShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetChronology(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestSubround_NilContainerHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetHasher(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubround_NilContainerMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetMarshalizer(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubround_NilContainerMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetMultiSignerContainer(cryptoMocks.NewMultiSignerContainerMock(nil))

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubround_NilContainerRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetRoundHandler(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubround_NilContainerShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetShardCoordinator(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubround_NilContainerSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetSyncTimer(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubround_NilContainerValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetNodesCoordinator(nil)

	sr, err := spos.NewSubround(
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

	assert.Nil(t, sr)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestSubround_EmptyChainIDShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	sr, err := spos.NewSubround(
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
		nil,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	assert.Equal(t, spos.ErrInvalidChainID, err)
	assert.Nil(t, sr)
}

func TestSubround_NewSubroundShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	sr, err := spos.NewSubround(
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

	assert.Nil(t, err)

	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.NotNil(t, sr)
}

func TestSubround_DoWorkShouldReturnFalseWhenJobFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
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
	sr.Job = nil
	sr.Check = func() bool {
		return true
	}

	maxTime := time.Now().Add(100 * time.Millisecond)
	roundHandlerMock := &consensus.RoundHandlerMock{}
	roundHandlerMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Until(maxTime)
	}

	r := sr.DoWork(context.Background(), roundHandlerMock)

	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenCheckFunctionIsNotSet(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
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
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = nil

	maxTime := time.Now().Add(100 * time.Millisecond)
	roundHandlerMock := &consensus.RoundHandlerMock{}
	roundHandlerMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Until(maxTime)
	}

	r := sr.DoWork(context.Background(), roundHandlerMock)
	assert.False(t, r)
}

func TestSubround_DoWorkShouldReturnFalseWhenConsensusIsNotDone(t *testing.T) {
	t.Parallel()

	testDoWork(t, false, false)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobAndConsensusAreDone(t *testing.T) {
	t.Parallel()

	testDoWork(t, true, true)
}

func testDoWork(t *testing.T, checkDone bool, shouldWork bool) {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
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
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return checkDone
	}

	maxTime := time.Now().Add(100 * time.Millisecond)
	roundHandlerMock := &consensus.RoundHandlerMock{}
	roundHandlerMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Until(maxTime)
	}

	r := sr.DoWork(context.Background(), roundHandlerMock)
	assert.Equal(t, shouldWork, r)
}

func TestSubround_DoWorkShouldReturnTrueWhenJobIsDoneAndConsensusIsDoneAfterAWhile(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
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

	var mut sync.RWMutex
	mut.Lock()
	checkSuccess := false
	mut.Unlock()

	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		mut.RLock()
		defer mut.RUnlock()
		return checkSuccess
	}

	maxTime := time.Now().Add(2000 * time.Millisecond)
	roundHandlerMock := &consensus.RoundHandlerMock{}
	roundHandlerMock.RemainingTimeCalled = func(time.Time, time.Duration) time.Duration {
		return time.Until(maxTime)
	}

	go func() {
		time.Sleep(1000 * time.Millisecond)

		mut.Lock()
		checkSuccess = true
		mut.Unlock()

		ch <- true
	}()

	r := sr.DoWork(context.Background(), roundHandlerMock)

	assert.True(t, r)
}

func TestSubround_Previous(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, bls.SrStartRound, sr.Previous())
}

func TestSubround_Current(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, bls.SrBlock, sr.Current())
}

func TestSubround_Next(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, bls.SrSignature, sr.Next())
}

func TestSubround_StartTime(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetRoundHandler(initRoundHandlerMock())
	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(25*roundTimeDuration/100),
		int64(40*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.StartTime())
}

func TestSubround_EndTime(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()
	container.SetRoundHandler(initRoundHandlerMock())
	sr, _ := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, int64(25*roundTimeDuration/100), sr.EndTime())
}

func TestSubround_Name(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	sr.Job = func(_ context.Context) bool {
		return true
	}
	sr.Check = func() bool {
		return false
	}

	assert.Equal(t, "(BLOCK)", sr.Name())
}

func TestSubround_GetAssociatedPid(t *testing.T) {
	t.Parallel()

	keysHandler := &testscommon.KeysHandlerStub{}
	consensusState := internalInitConsensusStateWithKeysHandler(keysHandler)
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	subround, _ := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	wasCalled := false
	pid := core.PeerID("a pid")
	providedPkBytes := []byte("pk bytes")
	keysHandler.GetAssociatedPidCalled = func(pkBytes []byte) core.PeerID {
		assert.Equal(t, providedPkBytes, pkBytes)
		wasCalled = true
		return pid
	}

	assert.Equal(t, pid, subround.GetAssociatedPid(providedPkBytes))
	assert.True(t, wasCalled)
}

func TestSubround_ShouldConsiderSelfKeyInConsensus(t *testing.T) {
	t.Parallel()

	t.Run("is main machine active, should return true", func(t *testing.T) {
		t.Parallel()

		consensusState := initConsensusState()
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		redundancyHandler := &mock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
			IsMainMachineActiveCalled: func() bool {
				return true
			},
		}
		container.SetNodeRedundancyHandler(redundancyHandler)

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		require.True(t, sr.ShouldConsiderSelfKeyInConsensus())
	})

	t.Run("is redundancy node machine active, should return true", func(t *testing.T) {
		t.Parallel()

		consensusState := initConsensusState()
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		redundancyHandler := &mock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return true
			},
			IsMainMachineActiveCalled: func() bool {
				return false
			},
		}
		container.SetNodeRedundancyHandler(redundancyHandler)

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		require.True(t, sr.ShouldConsiderSelfKeyInConsensus())
	})

	t.Run("is redundancy node machine but inactive, should return false", func(t *testing.T) {
		t.Parallel()

		consensusState := initConsensusState()
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		redundancyHandler := &mock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return true
			},
			IsMainMachineActiveCalled: func() bool {
				return true
			},
		}
		container.SetNodeRedundancyHandler(redundancyHandler)

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		require.False(t, sr.ShouldConsiderSelfKeyInConsensus())
	})
}

func TestSubround_GetLeaderStartRoundMessage(t *testing.T) {
	t.Parallel()

	t.Run("should work with multi key node", func(t *testing.T) {
		t.Parallel()

		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return bytes.Equal([]byte("1"), pkBytes)
			},
		}
		consensusState := internalInitConsensusStateWithKeysHandler(keysHandler)
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetSelfPubKey("1")

		require.Equal(t, spos.LeaderMultiKeyStartMsg, sr.GetLeaderStartRoundMessage())
	})

	t.Run("should work with single key node", func(t *testing.T) {
		t.Parallel()

		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return bytes.Equal([]byte("2"), pkBytes)
			},
		}
		consensusState := internalInitConsensusStateWithKeysHandler(keysHandler)
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetSelfPubKey("1")

		require.Equal(t, spos.LeaderSingleKeyStartMsg, sr.GetLeaderStartRoundMessage())
	})

	t.Run("should return empty string when leader is not managed by current node", func(t *testing.T) {
		t.Parallel()

		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return false
			},
		}
		consensusState := internalInitConsensusStateWithKeysHandler(keysHandler)
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetSelfPubKey("5")

		require.Equal(t, "", sr.GetLeaderStartRoundMessage())
	})
}

func TestSubround_IsSelfInConsensusGroup(t *testing.T) {
	t.Parallel()

	t.Run("should work with multi key node", func(t *testing.T) {
		t.Parallel()

		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return bytes.Equal([]byte("1"), pkBytes)
			},
		}
		consensusState := internalInitConsensusStateWithKeysHandler(keysHandler)
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		require.True(t, sr.IsSelfInConsensusGroup())
	})

	t.Run("should work with single key node", func(t *testing.T) {
		t.Parallel()

		consensusState := internalInitConsensusStateWithKeysHandler(&testscommon.KeysHandlerStub{})
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetSelfPubKey("1")

		require.True(t, sr.IsSelfInConsensusGroup())
	})
}

func TestSubround_IsSelfLeader(t *testing.T) {
	t.Parallel()

	t.Run("should work with multi key node", func(t *testing.T) {
		t.Parallel()

		keysHandler := &testscommon.KeysHandlerStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return bytes.Equal([]byte("1"), pkBytes)
			},
		}
		consensusState := internalInitConsensusStateWithKeysHandler(keysHandler)
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		sr.SetLeader("1")

		require.True(t, sr.IsSelfLeader())
	})

	t.Run("should work with single key node", func(t *testing.T) {
		t.Parallel()

		consensusState := internalInitConsensusStateWithKeysHandler(&testscommon.KeysHandlerStub{})
		ch := make(chan bool, 1)
		container := consensus.InitConsensusCore()

		sr, _ := spos.NewSubround(
			bls.SrStartRound,
			bls.SrBlock,
			bls.SrSignature,
			int64(5*roundTimeDuration/100),
			int64(25*roundTimeDuration/100),
			"(BLOCK)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetSelfPubKey("1")
		sr.SetLeader("1")

		require.True(t, sr.IsSelfLeader())
	})
}

func TestSubround_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var sr *spos.Subround
	require.True(t, sr.IsInterfaceNil())

	consensusState := internalInitConsensusStateWithKeysHandler(&testscommon.KeysHandlerStub{})
	ch := make(chan bool, 1)
	container := consensus.InitConsensusCore()

	sr, _ = spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	require.False(t, sr.IsInterfaceNil())
}
