package bls

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
	cnsTest "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignSubRoundEndOutGoingTxData(t *testing.T) {
	t.Parallel()

	t.Run("nil signing handler, should return error", func(t *testing.T) {
		sovSigHandler, err := NewSovereignSubRoundSignatureOutGoingTxData(nil)
		require.Equal(t, spos.ErrNilSigningHandler, err)
		require.True(t, check.IfNil(sovSigHandler))
	})

	t.Run("should work", func(t *testing.T) {
		sovSigHandler, err := NewSovereignSubRoundSignatureOutGoingTxData(&cnsTest.SigningHandlerStub{})
		require.Nil(t, err)
		require.False(t, sovSigHandler.IsInterfaceNil())
	})
}

func TestSovereignSubRoundSignatureOutGoingTxData_CreateSignatureShare(t *testing.T) {
	t.Parallel()

	outGoingOpHash := []byte("outGoingOpHash")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
			Epoch: 3,
		},
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			OutGoingOperationsHash: outGoingOpHash,
		},
	}
	selfPubKey := []byte("pubKey")
	selfIndex := uint16(4)

	expectedSigShare := []byte("sigShare")
	createSigShareCt := 0
	signingHandler := &cnsTest.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			require.Equal(t, outGoingOpHash, message)
			require.Equal(t, selfIndex, index)
			require.Equal(t, selfPubKey, publicKeyBytes)
			require.Equal(t, sovHdr.GetEpoch(), epoch)

			createSigShareCt++
			return expectedSigShare, nil
		},
	}
	sovSigHandler, _ := NewSovereignSubRoundSignatureOutGoingTxData(signingHandler)

	t.Run("invalid header type, should return error", func(t *testing.T) {
		sigShare, err := sovSigHandler.CreateSignatureShare(sovHdr.Header, selfIndex, selfPubKey)
		require.Nil(t, sigShare)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeader = nil

		sigShare, err := sovSigHandler.CreateSignatureShare(&sovHdrCopy, selfIndex, selfPubKey)
		require.Empty(t, sigShare)
		require.Nil(t, err)
	})

	t.Run("should create sig share", func(t *testing.T) {
		sigShare, err := sovSigHandler.CreateSignatureShare(sovHdr, selfIndex, selfPubKey)
		require.Equal(t, expectedSigShare, sigShare)
		require.Nil(t, err)
		require.Equal(t, 1, createSigShareCt)
	})
}

func TestSovereignSubRoundSignatureOutGoingTxData_AddSigShareToConsensusMessage(t *testing.T) {
	t.Parallel()

	cnsMsg := &consensus.Message{
		SignatureShare: []byte("sigShare"),
	}

	sovSigHandler, _ := NewSovereignSubRoundSignatureOutGoingTxData(&cnsTest.SigningHandlerStub{})
	sovSigHandler.AddSigShareToConsensusMessage([]byte("sigShareOutGoingTxData"), cnsMsg)
	require.Equal(t, &consensus.Message{
		SignatureShare:               []byte("sigShare"),
		SignatureShareOutGoingTxData: []byte("sigShareOutGoingTxData"),
	}, cnsMsg)
}

func TestSovereignSubRoundSignatureOutGoingTxData_StoreSignatureShare(t *testing.T) {
	t.Parallel()

	cnsMsg := &consensus.Message{
		SignatureShare:               []byte("sigShare"),
		SignatureShareOutGoingTxData: []byte("sigShareOutGoingTxData"),
	}

	expectedIdx := uint16(4)
	wasSigStored := false
	signHandler := &cnsTest.SigningHandlerStub{
		StoreSignatureShareCalled: func(index uint16, sig []byte) error {
			require.Equal(t, expectedIdx, index)
			require.Equal(t, cnsMsg.SignatureShareOutGoingTxData, sig)

			wasSigStored = true
			return nil
		},
	}

	sovSigHandler, _ := NewSovereignSubRoundSignatureOutGoingTxData(signHandler)
	err := sovSigHandler.StoreSignatureShare(expectedIdx, cnsMsg)
	require.Nil(t, err)
	require.True(t, wasSigStored)
}

func TestSovereignSubRoundSignatureOutGoingTxData_Identifier(t *testing.T) {
	t.Parallel()

	sovSigHandler, _ := NewSovereignSubRoundSignatureOutGoingTxData(&cnsTest.SigningHandlerStub{})
	require.Equal(t, "sovereignSubRoundSignatureOutGoingTxData", sovSigHandler.Identifier())
}
