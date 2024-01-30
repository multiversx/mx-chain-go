package bls_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	testscommonOutport "github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/subRoundsHolder"
	"github.com/stretchr/testify/assert"
)

var chainID = []byte("chain ID")

const currentPid = core.PeerID("pid")

const roundTimeDuration = 100 * time.Millisecond

func displayStatistics() {
}

func extend(subroundId int) {
	fmt.Println(subroundId)
}

// executeStoredMessages tries to execute all the messages received which are valid for execution
func executeStoredMessages() {
}

// resetConsensusMessages resets at the start of each round, all the previous consensus messages received
func resetConsensusMessages() {
}

func initRoundHandlerMock() *mock.RoundHandlerMock {
	return &mock.RoundHandlerMock{
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
	sposWorker := &mock.SposWorkerMock{}
	sposWorker.GetConsensusStateChangedChannelsCalled = func() chan bool {
		return make(chan bool)
	}
	sposWorker.RemoveAllReceivedMessagesCallsCalled = func() {}

	sposWorker.AddReceivedMessageCallCalled =
		func(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
		}

	return sposWorker
}

func initFactoryWithContainer(container *mock.ConsensusCoreMock) bls.Factory {
	worker := initWorker()
	consensusState := initConsensusState()

	fct, _ := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	return fct
}

func initFactory() bls.Factory {
	container := mock.InitConsensusCore()
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

	consensusState := initConsensusState()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		nil,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilConsensusCore, err)
}

func TestFactory_NewFactoryNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		nil,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestFactory_NewFactoryNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetBlockchain(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestFactory_NewFactoryNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetBlockProcessor(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestFactory_NewFactoryNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetBootStrapper(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestFactory_NewFactoryNilChronologyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetChronology(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestFactory_NewFactoryNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetHasher(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestFactory_NewFactoryNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetMarshalizer(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestFactory_NewFactoryNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetMultiSignerContainer(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestFactory_NewFactoryNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetRoundHandler(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestFactory_NewFactoryNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetShardCoordinator(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestFactory_NewFactoryNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetSyncTimer(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_NewFactoryNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()
	container.SetValidatorGroupSelector(nil)

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestFactory_NewFactoryNilWorkerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		nil,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilWorker, err)
}

func TestFactory_NewFactoryNilAppStatusHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		nil,
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilAppStatusHandler, err)
}

func TestFactory_NewFactoryNilSignaturesTrackerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		nil,
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilSentSignatureTracker, err)
}

func TestFactory_NewFactoryNilEnableEpochHandlerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		nil,
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilEnableEpochHandler, err)
}

func TestFactory_NewFactoryNilExtraSignersHolderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		nil,
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, errors.ErrNilExtraSignersHolder, err)
}

func TestFactory_NewFactoryNilSubRoundEndV2CreatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		nil,
	)

	assert.Nil(t, fct)
	assert.Equal(t, errors.ErrNilSubRoundEndV2Creator, err)
}

func TestFactory_NewFactoryShouldWork(t *testing.T) {
	t.Parallel()

	fct := *initFactory()

	assert.False(t, check.IfNil(&fct))
}

func TestFactory_NewFactoryEmptyChainIDShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	container := mock.InitConsensusCore()
	worker := initWorker()

	fct, err := bls.NewSubroundsFactory(
		container,
		consensusState,
		worker,
		nil,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		consensus.ConsensusModelV1,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)

	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrInvalidChainID, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*mock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	err := fct.GenerateStartRoundSubround()

	assert.Equal(t, spos.ErrNilChannel, err)
}

func TestFactory_GenerateSubroundStartRoundShouldFailWhenNewSubroundStartRoundFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)
	container.SetSyncTimer(nil)

	err := fct.GenerateStartRoundSubround()

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestFactory_GenerateSubroundBlock(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		fct := *initFactory()
		fct.Worker().(*mock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		err := fct.GenerateBlockSubroundV1()
		assert.NotNil(t, err)

		err = fct.GenerateBlockSubroundV2()
		assert.NotNil(t, err)
	})
	t.Run("should work with v1", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateBlockSubroundV1()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundBlock", fmt.Sprintf("%T", addedSubround))
	})
	t.Run("should work with v2", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		worker := initWorker()
		consensusState := initConsensusState()

		fct, _ := bls.NewSubroundsFactory(
			container,
			consensusState,
			worker,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
			&mock.SentSignatureTrackerStub{},
			consensus.ConsensusModelV2,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&subRoundsHolder.ExtraSignersHolderMock{},
			bls.NewSubRoundEndV2Creator(),
		)

		err := fct.GenerateBlockSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundBlockV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundSignature(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		fct := *initFactory()
		fct.Worker().(*mock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		err := fct.GenerateSignatureSubroundV1()
		assert.Equal(t, spos.ErrNilChannel, err)

		err = fct.GenerateSignatureSubroundV2()
		assert.Equal(t, spos.ErrNilChannel, err)
	})
	t.Run("should fail when new subround Signature fails", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		fct := *initFactoryWithContainer(container)
		container.SetSyncTimer(nil)

		err := fct.GenerateSignatureSubroundV1()
		assert.Equal(t, spos.ErrNilSyncTimer, err)

		err = fct.GenerateSignatureSubroundV2()
		assert.Equal(t, spos.ErrNilSyncTimer, err)
	})
	t.Run("should work with v1", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateSignatureSubroundV1()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundSignature", fmt.Sprintf("%T", addedSubround))
	})
	t.Run("should work with v2", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		worker := initWorker()
		consensusState := initConsensusState()

		fct, _ := bls.NewSubroundsFactory(
			container,
			consensusState,
			worker,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
			&mock.SentSignatureTrackerStub{},
			consensus.ConsensusModelV2,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&subRoundsHolder.ExtraSignersHolderMock{},
			bls.NewSubRoundEndV2Creator(),
		)

		err := fct.GenerateSignatureSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundSignatureV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundEndRound(t *testing.T) {
	t.Parallel()

	t.Run("should fail when new subround fails", func(t *testing.T) {
		t.Parallel()

		fct := *initFactory()
		fct.Worker().(*mock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
			return nil
		}

		err := fct.GenerateEndRoundSubroundV1()
		assert.Equal(t, spos.ErrNilChannel, err)

		err = fct.GenerateEndRoundSubroundV2()
		assert.Equal(t, spos.ErrNilChannel, err)
	})
	t.Run("should fail when new subround EndRound fails", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		fct := *initFactoryWithContainer(container)
		container.SetSyncTimer(nil)

		err := fct.GenerateEndRoundSubroundV1()
		assert.Equal(t, spos.ErrNilSyncTimer, err)

		err = fct.GenerateEndRoundSubroundV2()
		assert.Equal(t, spos.ErrNilSyncTimer, err)
	})
	t.Run("should work with v1", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		fct := *initFactoryWithContainer(container)

		err := fct.GenerateEndRoundSubroundV1()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundEndRound", fmt.Sprintf("%T", addedSubround))
	})
	t.Run("should work with v2", func(t *testing.T) {
		t.Parallel()

		var addedSubround consensus.SubroundHandler
		container := mock.InitConsensusCore()
		container.SetChronology(&mock.ChronologyHandlerMock{
			AddSubroundCalled: func(handler consensus.SubroundHandler) {
				addedSubround = handler
			},
		})
		worker := initWorker()
		consensusState := initConsensusState()

		fct, _ := bls.NewSubroundsFactory(
			container,
			consensusState,
			worker,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
			&mock.SentSignatureTrackerStub{},
			consensus.ConsensusModelV2,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&subRoundsHolder.ExtraSignersHolderMock{},
			bls.NewSubRoundEndV2Creator(),
		)

		err := fct.GenerateEndRoundSubroundV2()
		assert.Nil(t, err)
		assert.Equal(t, "*bls.subroundEndRoundV2", fmt.Sprintf("%T", addedSubround))
	})
}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	t.Parallel()

	subroundHandlers := 0

	chrm := &mock.ChronologyHandlerMock{}
	chrm.AddSubroundCalled = func(subroundHandler consensus.SubroundHandler) {
		subroundHandlers++
	}
	container := mock.InitConsensusCore()
	container.SetChronology(chrm)
	fct := *initFactoryWithContainer(container)
	fct.SetOutportHandler(&testscommonOutport.OutportStub{})

	err := fct.GenerateSubrounds()
	assert.Nil(t, err)

	assert.Equal(t, 4, subroundHandlers)
}

func TestFactory_GenerateSubroundsNilOutportShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	err := fct.GenerateSubrounds()
	assert.Equal(t, outport.ErrNilDriver, err)
}

func TestFactory_GenerateSubroundsInvalidConsensusModelShouldFail(t *testing.T) {
	t.Parallel()

	worker := initWorker()
	consensusState := initConsensusState()

	fct, _ := bls.NewSubroundsFactory(
		mock.InitConsensusCore(),
		consensusState,
		worker,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&mock.SentSignatureTrackerStub{},
		"invalid",
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&subRoundsHolder.ExtraSignersHolderMock{},
		bls.NewSubRoundEndV2Creator(),
	)
	fct.SetOutportHandler(&testscommonOutport.OutportStub{})

	err := fct.GenerateSubrounds()
	assert.ErrorIs(t, err, errors.ErrUnimplementedConsensusModel)
}

func TestFactory_SetIndexerShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	outportHandler := &testscommonOutport.OutportStub{}
	fct.SetOutportHandler(outportHandler)

	assert.Equal(t, outportHandler, fct.Outport())
}
