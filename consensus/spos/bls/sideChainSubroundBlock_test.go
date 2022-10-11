package bls_test

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSideChainSubroundBlock_DoBlockJob(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	scsr, _ := bls.NewSideChainSubroundBlock(sr)
	r := scsr.DoBlockJob()
	assert.False(t, r)

	scsr.SetSelfPubKey(scsr.ConsensusGroup()[0])
	_ = scsr.SetJobDone(scsr.SelfPubKey(), bls.SrBlock, true)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	container.SetRoundHandler(&mock.RoundHandlerMock{
		RoundIndex: 1,
	})
	scsr.SetStatus(bls.SrBlock, spos.SsFinished)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	_ = scsr.SetJobDone(scsr.SelfPubKey(), bls.SrBlock, false)
	bpm := &mock.BlockProcessorMock{}
	err := errors.New("error")
	bpm.CreateBlockCalled = func(header data.HeaderHandler, remainingTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		return header, nil, err
	}
	container.SetBlockProcessor(bpm)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	scsr.SetStatus(bls.SrBlock, spos.SsNotFinished)
	bpm = mock.InitBlockProcessorMock()
	bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
		return nil, err
	}
	container.SetBlockProcessor(bpm)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	bpm = mock.InitBlockProcessorMock()
	bpm.CreateBlockCalled = func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		return nil, nil, err
	}
	container.SetBlockProcessor(bpm)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	bpm = mock.InitBlockProcessorMock()
	container.SetBlockProcessor(bpm)
	bm := &mock.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			return err
		},
	}
	container.SetBroadcastMessenger(bm)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	bpm = mock.InitBlockProcessorMock()
	bpm.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
		return nil, nil, err
	}
	container.SetBlockProcessor(bpm)
	bm = &mock.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			return nil
		},
	}
	container.SetBroadcastMessenger(bm)
	r = scsr.DoBlockJob()
	assert.False(t, r)

	bpm = mock.InitBlockProcessorMock()
	container.SetBlockProcessor(bpm)
	r = scsr.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), sr.Header.GetNonce())
}
