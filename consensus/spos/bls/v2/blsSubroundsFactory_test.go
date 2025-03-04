package v2_test

import (
	"context"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonConsensus "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	testscommonOutport "github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

var chainID = []byte("chain ID")

const currentPid = core.PeerID("pid")

const roundTimeDuration = 100 * time.Millisecond

// executeStoredMessages tries to execute all the messages received which are valid for execution
func executeStoredMessages() {
}

func initRoundHandlerMock() *testscommonConsensus.RoundHandlerMock {
	return &testscommonConsensus.RoundHandlerMock{
		RoundIndex: 0,
		TimeStampCalled: func() time.Time {
			return time.Unix(0, 0)
		},
		TimeDurationCalled: func() time.Duration {
			return roundTimeDuration
		},
	}
}

func initWorker() spos.WorkerHandler {
	sposWorker := &testscommonConsensus.SposWorkerMock{}
	sposWorker.GetConsensusStateChangedChannelsCalled = func() chan bool {
		return make(chan bool)
	}
	sposWorker.RemoveAllReceivedMessagesCallsCalled = func() {}

	sposWorker.AddReceivedMessageCallCalled =
		func(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
		}

	return sposWorker
}

func initFactoryWithContainer(container *testscommonConsensus.ConsensusCoreMock) v2.Factory {
	worker := initWorker()
	consensusState := initializers.InitConsensusState()

	fct, _ := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	return fct
}

func initFactory() v2.Factory {
	container := testscommonConsensus.InitConsensusCore()
	return initFactoryWithContainer(container)
}

func TestFactory_GetMessageTypeName(t *testing.T) {
	t.Parallel()

	r := bls.GetStringValue(bls.MtBlockBodyAndHeader)
	assert.Equal(t, "(BLOCK_BODY_AND_HEADER)", r)

	r = bls.GetStringValue(bls.MtBlockBody)
	assert.Equal(t, "(BLOCK_BODY)", r)

	r = bls.GetStringValue(bls.MtBlockHeader)
	assert.Equal(t, "(BLOCK_HEADER)", r)

	r = bls.GetStringValue(bls.MtSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bls.GetStringValue(bls.MtBlockHeaderFinalInfo)
	assert.Equal(t, "(FINAL_INFO)", r)

	r = bls.GetStringValue(bls.MtUnknown)
	assert.Equal(t, "(UNKNOWN)", r)

	r = bls.GetStringValue(consensus.MessageType(-1))
	assert.Equal(t, "Undefined message type", r)
}

func TestFactory_NewFactoryNilContainerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	worker := initWorker()

	fct, err := v2.NewSubroundsFactory(
		nil,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilConsensusCore, err)
}

func TestFactory_NewFactoryNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()

	fct, err := v2.NewSubroundsFactory(
		container,
		nil,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestFactory_NewFactoryNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetBlockchain(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestFactory_NewFactoryNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetBlockProcessor(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestFactory_NewFactoryNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetBootStrapper(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestFactory_NewFactoryNilChronologyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetChronology(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestFactory_NewFactoryNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetHasher(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestFactory_NewFactoryNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetMarshalizer(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestFactory_NewFactoryNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetMultiSignerContainer(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestFactory_NewFactoryNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetRoundHandler(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestFactory_NewFactoryNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetShardCoordinator(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestFactory_NewFactoryNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetSyncTimer(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_NewFactoryNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()
	container.SetValidatorGroupSelector(nil)

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestFactory_NewFactoryNilWorkerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		nil,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilWorker, err)
}

func TestFactory_NewFactoryNilAppStatusHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		nil,
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilAppStatusHandler, err)
}

func TestFactory_NewFactoryNilSignaturesTrackerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		nil,
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, v2.ErrNilSentSignatureTracker, err)
}

func TestFactory_NewFactoryNilThrottlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		nil,
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilThrottler, err)
}

func TestFactory_NewFactoryShouldWork(t *testing.T) {
	t.Parallel()

	fct := *initFactory()

	assert.False(t, check.IfNil(&fct))
}

func TestFactory_NewFactoryEmptyChainIDShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initializers.InitConsensusState()
	container := testscommonConsensus.InitConsensusCore()
	worker := initWorker()

	fct, err := v2.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		nil,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&dataRetrieverMocks.ThrottlerStub{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrInvalidChainID, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*testscommonConsensus.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	err := fct.GenerateStartRoundSubround()

	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundStartRoundFail(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	fct := *initFactoryWithContainer(container)
	container.SetSyncTimer(nil)

	err := fct.GenerateStartRoundSubround()

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundBlockShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*testscommonConsensus.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	err := fct.GenerateBlockSubround()

	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundBlockShouldFailWhenNewSubroundBlockFail(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	fct := *initFactoryWithContainer(container)
	container.SetSyncTimer(nil)

	err := fct.GenerateBlockSubround()

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundSignatureShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*testscommonConsensus.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	err := fct.GenerateSignatureSubround()

	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundSignatureShouldFailWhenNewSubroundSignatureFail(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	fct := *initFactoryWithContainer(container)
	container.SetSyncTimer(nil)

	err := fct.GenerateSignatureSubround()

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundEndRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*testscommonConsensus.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	err := fct.GenerateEndRoundSubround()

	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundEndRoundShouldFailWhenNewSubroundEndRoundFail(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	fct := *initFactoryWithContainer(container)
	container.SetSyncTimer(nil)

	err := fct.GenerateEndRoundSubround()

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	t.Parallel()

	subroundHandlers := 0

	chrm := &testscommonConsensus.ChronologyHandlerMock{}
	chrm.AddSubroundCalled = func(subroundHandler consensus.SubroundHandler) {
		subroundHandlers++
	}
	container := testscommonConsensus.InitConsensusCore()
	container.SetChronology(chrm)
	fct := *initFactoryWithContainer(container)
	fct.SetOutportHandler(&testscommonOutport.OutportStub{})

	err := fct.GenerateSubrounds()
	assert.Nil(t, err)

	assert.Equal(t, 4, subroundHandlers)
}

func TestFactory_GenerateSubroundsNilOutportShouldFail(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	err := fct.GenerateSubrounds()
	assert.Equal(t, outport.ErrNilDriver, err)
}

func TestFactory_SetIndexerShouldWork(t *testing.T) {
	t.Parallel()

	container := testscommonConsensus.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	outportHandler := &testscommonOutport.OutportStub{}
	fct.SetOutportHandler(outportHandler)

	assert.Equal(t, outportHandler, fct.Outport())
}
