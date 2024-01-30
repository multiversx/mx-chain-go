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
		sovSigHandler, err := NewSovereignSubRoundEndOutGoingTxData(nil)
		require.Equal(t, spos.ErrNilSigningHandler, err)
		require.True(t, check.IfNil(sovSigHandler))
	})

	t.Run("should work", func(t *testing.T) {
		sovSigHandler, err := NewSovereignSubRoundEndOutGoingTxData(&cnsTest.SigningHandlerStub{})
		require.Nil(t, err)
		require.False(t, sovSigHandler.IsInterfaceNil())
	})
}

func TestSovereignSubRoundEndOutGoingTxData_VerifyAggregatedSignatures(t *testing.T) {
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

	verifyCalledCt := 0
	expectedBitMap := []byte{0x3}
	signingHandler := &cnsTest.SigningHandlerStub{
		VerifyCalled: func(msg []byte, bitmap []byte, epoch uint32) error {
			require.Equal(t, expectedBitMap, bitmap)
			require.Equal(t, outGoingOpHash, msg)
			require.Equal(t, sovHdr.GetEpoch(), epoch)

			verifyCalledCt++
			return nil
		},
	}
	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(signingHandler)

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovSigHandler.VerifyAggregatedSignatures(expectedBitMap, sovHdr.Header)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeader = nil
		err := sovSigHandler.VerifyAggregatedSignatures(expectedBitMap, &sovHdrCopy)
		require.Nil(t, err)
		require.Zero(t, verifyCalledCt)
	})

	t.Run("should create sig share", func(t *testing.T) {
		err := sovSigHandler.VerifyAggregatedSignatures(expectedBitMap, sovHdr)
		require.Nil(t, err)
		require.Equal(t, 1, verifyCalledCt)
	})
}

func TestSovereignSubRoundEndOutGoingTxData_AggregateSignatures(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(4)
	aggregateSigsCalledCt := 0
	setAggregateSigsCalledCt := 0
	expectedBitMap := []byte{0x3}
	aggregatedSig := []byte("aggregatedSig")

	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: expectedEpoch,
		},
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			OutGoingOperationsHash: []byte("hash"),
		},
	}

	signingHandler := &cnsTest.SigningHandlerStub{
		AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
			require.Equal(t, expectedBitMap, bitmap)
			require.Equal(t, expectedEpoch, epoch)

			aggregateSigsCalledCt++
			return aggregatedSig, nil
		},

		SetAggregatedSigCalled: func(sig []byte) error {
			require.Equal(t, aggregatedSig, sig)

			setAggregateSigsCalledCt++
			return nil
		},
	}
	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(signingHandler)
	result, err := sovSigHandler.AggregateAndSetSignatures(expectedBitMap, sovHdr)
	require.Nil(t, err)
	require.Equal(t, aggregatedSig, result)
}

func TestSovereignSubRoundEndOutGoingTxData_SeAggregatedSignatureInHeader(t *testing.T) {
	t.Parallel()

	aggregatedSig := []byte("aggregatedSig")
	outGoingOpHash := []byte("outGoingOpHash")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: 3,
		},
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			OutGoingOperationsHash: outGoingOpHash,
		},
	}

	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(&cnsTest.SigningHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovSigHandler.SetAggregatedSignatureInHeader(sovHdr.Header, aggregatedSig)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeader = nil
		err := sovSigHandler.SetAggregatedSignatureInHeader(&sovHdrCopy, aggregatedSig)
		require.Nil(t, err)
		require.True(t, check.IfNil(sovHdrCopy.OutGoingMiniBlockHeader))
	})

	t.Run("should add sig share", func(t *testing.T) {
		err := sovSigHandler.SetAggregatedSignatureInHeader(sovHdr, aggregatedSig)
		require.Nil(t, err)
		require.Equal(t, &block.SovereignChainHeader{
			Header: &block.Header{
				Epoch: 3,
			},
			OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: aggregatedSig,
			},
		}, sovHdr)
	})
}

func TestSovereignSubRoundEndOutGoingTxData_SignAndSetLeaderSignature(t *testing.T) {
	t.Parallel()

	outGoingOpHash := []byte("outGoingOpHash")
	aggregatedSig := []byte("aggregatedSig")
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
			Epoch: 3,
		},
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			OutGoingOperationsHash:                outGoingOpHash,
			AggregatedSignatureOutGoingOperations: aggregatedSig,
		},
	}

	verifyCalledCt := 0
	expectedLeaderPubKey := []byte("leaderPubKey")
	expectedLeaderSig := []byte("leaderSig")
	signingHandler := &cnsTest.SigningHandlerStub{
		CreateSignatureForPublicKeyCalled: func(message []byte, publicKeyBytes []byte) ([]byte, error) {
			require.Equal(t, expectedLeaderPubKey, publicKeyBytes)
			require.Equal(t, append(outGoingOpHash, aggregatedSig...), message)

			verifyCalledCt++
			return expectedLeaderSig, nil
		},
	}
	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(signingHandler)

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovSigHandler.SignAndSetLeaderSignature(sovHdr.Header, expectedLeaderPubKey)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeader = nil
		err := sovSigHandler.SignAndSetLeaderSignature(&sovHdrCopy, expectedLeaderPubKey)
		require.Nil(t, err)
		require.Zero(t, verifyCalledCt)
	})

	t.Run("should create leader sig", func(t *testing.T) {
		err := sovSigHandler.SignAndSetLeaderSignature(sovHdr, expectedLeaderPubKey)
		require.Nil(t, err)
		require.Equal(t, 1, verifyCalledCt)
		require.Equal(t, &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
				Epoch: 3,
			},
			OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: aggregatedSig,
				LeaderSignatureOutGoingOperations:     expectedLeaderSig,
			},
		}, sovHdr)
	})
}

func TestSovereignSubRoundEndOutGoingTxData_HaveConsensusHeaderWithFullInfo(t *testing.T) {
	t.Parallel()

	aggregatedSig := []byte("aggregatedSig")
	leaderSig := []byte("leaderSig")
	cnsMsg := &consensus.Message{
		AggregatedSignatureOutGoingTxData: aggregatedSig,
		LeaderSignatureOutGoingTxData:     leaderSig,
	}
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

	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(&cnsTest.SigningHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovSigHandler.SetConsensusDataInHeader(sovHdr.Header, cnsMsg)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeader = nil
		err := sovSigHandler.SetConsensusDataInHeader(&sovHdrCopy, cnsMsg)
		require.Nil(t, err)
		require.True(t, check.IfNil(sovHdrCopy.OutGoingMiniBlockHeader))
	})

	t.Run("should create leader sig", func(t *testing.T) {
		err := sovSigHandler.SetConsensusDataInHeader(sovHdr, cnsMsg)
		require.Nil(t, err)
		require.Equal(t, &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
				Epoch: 3,
			},
			OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
				OutGoingOperationsHash:                outGoingOpHash,
				AggregatedSignatureOutGoingOperations: aggregatedSig,
				LeaderSignatureOutGoingOperations:     leaderSig,
			},
		}, sovHdr)
	})
}

func TestSovereignSubRoundEndOutGoingTxData_AddLeaderAndAggregatedSignatures(t *testing.T) {
	t.Parallel()

	aggregatedSig := []byte("aggregatedSig")
	leaderSig := []byte("leaderSig")
	cnsMsg := &consensus.Message{}
	outGoingOpHash := []byte("outGoingOpHash")

	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
			Epoch: 3,
		},
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			OutGoingOperationsHash:                outGoingOpHash,
			AggregatedSignatureOutGoingOperations: aggregatedSig,
			LeaderSignatureOutGoingOperations:     leaderSig,
		},
	}

	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(&cnsTest.SigningHandlerStub{})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		err := sovSigHandler.AddLeaderAndAggregatedSignatures(sovHdr.Header, cnsMsg)
		require.ErrorIs(t, err, errors.ErrWrongTypeAssertion)
	})

	t.Run("no outgoing mini block header", func(t *testing.T) {
		sovHdrCopy := *sovHdr
		sovHdrCopy.OutGoingMiniBlockHeader = nil
		err := sovSigHandler.AddLeaderAndAggregatedSignatures(&sovHdrCopy, cnsMsg)
		require.Nil(t, err)
		require.True(t, check.IfNil(sovHdrCopy.OutGoingMiniBlockHeader))
	})

	t.Run("should add sigs", func(t *testing.T) {
		err := sovSigHandler.AddLeaderAndAggregatedSignatures(sovHdr, cnsMsg)
		require.Nil(t, err)
		require.Equal(t, &consensus.Message{
			AggregatedSignatureOutGoingTxData: aggregatedSig,
			LeaderSignatureOutGoingTxData:     leaderSig,
		}, cnsMsg)
	})
}

func TestSovereignSubRoundEndOutGoingTxData_Identifier(t *testing.T) {
	t.Parallel()
	sovSigHandler, _ := NewSovereignSubRoundEndOutGoingTxData(&cnsTest.SigningHandlerStub{})
	require.Equal(t, "sovereignSubRoundEndOutGoingTxData", sovSigHandler.Identifier())
}
