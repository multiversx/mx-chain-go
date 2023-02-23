package bls_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewSubroundBlockV2_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	srV2, err := bls.NewSubroundBlockV2(nil)
	assert.Nil(t, srV2)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSubroundBlockV2_ShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	srV2, err := bls.NewSubroundBlockV2(sr)
	assert.NotNil(t, srV2)
	assert.Nil(t, err)
}

func TestSubroundBlockV2_DoBlockJob(t *testing.T) {
	t.Parallel()

	t.Run("if self is not leader in current round, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if current round is less or equal than the round in last committed block, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		_ = srV2.SetJobDone(srV2.SelfPubKey(), bls.SrBlock, true)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if self job is done, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		_ = srV2.SetJobDone(srV2.SelfPubKey(), bls.SrBlock, true)
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		srV2.SetStatus(bls.SrBlock, spos.SsFinished)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if subround is finished, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		srV2.SetStatus(bls.SrBlock, spos.SsFinished)
		bpm := &testscommon.BlockProcessorStub{}
		err := errors.New("error")
		bpm.CreateBlockCalled = func(header data.HeaderHandler, remainingTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return header, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if create header fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if create block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		bpm.CreateBlockCalled = func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return nil, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if send block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		container.SetBlockProcessor(bpm)
		bm := &mock.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return err
			},
		}
		container.SetBroadcastMessenger(bm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if process received block fails, should return false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		err := errors.New("error")
		bpm.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
			return nil, nil, err
		}
		container.SetBlockProcessor(bpm)

		r := srV2.DoBlockJob()
		assert.False(t, r)
	})

	t.Run("if process received block succeeds, should return true", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundBlockV2(sr)
		srV2.SetSelfPubKey(srV2.ConsensusGroup()[0])
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})

		r := srV2.DoBlockJob()
		assert.True(t, r)
		assert.Equal(t, uint64(1), sr.Header.GetNonce())
	})
}
