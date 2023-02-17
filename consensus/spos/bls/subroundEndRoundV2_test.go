package bls_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewSubroundEndRoundV2_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	srV2, err := bls.NewSubroundEndRoundV2(nil)

	assert.Nil(t, srV2)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSubroundEndRoundV2_ShouldWork(t *testing.T) {
	t.Parallel()

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	srV2, err := bls.NewSubroundEndRoundV2(sr)

	assert.NotNil(t, srV2)
	assert.Nil(t, err)
}

func TestSubroundEndRoundV2_GetMessageToVerifySig(t *testing.T) {
	t.Parallel()

	t.Run("getMessageToVerifySig should return nil when CalculateHash method fails", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundEndRoundV2(sr)

		srV2.Header = nil
		msg := srV2.GetMessageToVerifySig()

		assert.Nil(t, msg)
	})

	t.Run("getMessageToVerifySig should return the message on which the signature should be verified", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
		srV2, _ := bls.NewSubroundEndRoundV2(sr)

		srV2.Header = &block.Header{Nonce: 1}
		expectedMsg, _ := core.CalculateHash(srV2.Marshalizer(), srV2.Hasher(), srV2.Header)
		msg := srV2.GetMessageToVerifySig()

		assert.Equal(t, expectedMsg, msg)
	})
}
