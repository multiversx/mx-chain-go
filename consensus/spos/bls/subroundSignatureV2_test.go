package bls_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/stretchr/testify/assert"
)

func TestNewSubroundSignatureV2_ShouldErrNilSubround(t *testing.T) {
	t.Parallel()

	srV2, err := bls.NewSubroundSignatureV2(nil)

	assert.Nil(t, srV2)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestNewSubroundSignatureV2_ShouldWork(t *testing.T) {
	t.Parallel()

	sr := initSubroundSignature()
	srV2, err := bls.NewSubroundSignatureV2(sr)

	assert.NotNil(t, srV2)
	assert.Nil(t, err)
}

func TestSubroundSignatureV2_GetMessageToSign(t *testing.T) {
	t.Parallel()

	t.Run("getMessageToSign should return nil when CalculateHash method fails", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		srV2, _ := bls.NewSubroundSignatureV2(sr)

		srV2.Header = nil
		msg := srV2.GetMessageToSign()

		assert.Nil(t, msg)
	})

	t.Run("getMessageToSign should return the message that should be signed", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		srV2, _ := bls.NewSubroundSignatureV2(sr)

		srV2.Header = &block.Header{Nonce: 1}
		expectedMsg, _ := core.CalculateHash(srV2.Marshalizer(), srV2.Hasher(), srV2.Header)
		msg := srV2.GetMessageToSign()

		assert.Equal(t, expectedMsg, msg)
	})
}
