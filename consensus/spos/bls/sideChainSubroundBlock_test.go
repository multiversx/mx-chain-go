package bls_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewSideChainSubroundBlock_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	scsr, err := bls.NewSideChainSubroundBlock(nil)
	assert.Nil(t, scsr)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSideChainSubroundBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	scsr, err := bls.NewSideChainSubroundBlock(sr)
	assert.NotNil(t, scsr)
	assert.Nil(t, err)
}

func TestSideChainSubroundBlock_DoBlockJob(t *testing.T) {
	t.Parallel()

	t.Run("if self is not leader in current round, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if current round is less or equal than the round in last committed block, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		_ = scsr.SetJobDone(scsr.SelfPubKey(), bls.SrBlock, true)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if self job is done, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		_ = scsr.SetJobDone(scsr.SelfPubKey(), bls.SrBlock, true)
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		scsr.SetStatus(bls.SrBlock, spos.SsFinished)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if subround is finished, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		scsr.SetStatus(bls.SrBlock, spos.SsFinished)
		bpm := &mock.BlockProcessorMock{}
		err := errors.New("error")
		bpm.CreateBlockCalled = func(header data.HeaderHandler, remainingTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return header, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if create header fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock()
		err := errors.New("error")
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return nil, err
		}
		container.SetBlockProcessor(bpm)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if create block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock()
		err := errors.New("error")
		bpm.CreateBlockCalled = func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return nil, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if send block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock()
		err := errors.New("error")
		container.SetBlockProcessor(bpm)
		bm := &mock.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return err
			},
		}
		container.SetBroadcastMessenger(bm)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if process received block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock()
		err := errors.New("error")
		bpm.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
			return nil, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := scsr.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if process received block succeeds, should return true", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		scsr, _ := bls.NewSideChainSubroundBlock(sr)
		scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})

		r := scsr.DoBlockJob()
		assert.True(t, r)
		assert.Equal(t, uint64(1), sr.Header.GetNonce())
	})
}
