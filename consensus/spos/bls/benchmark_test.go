package bls_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func BenchmarkSubroundEndRound_VerifyNodesOnAggSigFailTime(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer400Sig(container, &statusHandler.AppStatusHandlerMock{})

		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				return nil
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.Header = &block.Header{}
		for i := 0; i < len(sr.ConsensusGroup()); i++ {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
		}
		b.StartTimer()
		invalidSigners, err := sr.VerifyNodesOnAggSigFail()
		b.StopTimer()
		require.Nil(b, err)
		require.NotNil(b, invalidSigners)
	}
}

func BenchmarkSubroundEndRound_VerifyNodesOnAggSigFailTimeParallelNoThrottle(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer400Sig(container, &statusHandler.AppStatusHandlerMock{})

		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				return nil
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.Header = &block.Header{}
		for i := 0; i < len(sr.ConsensusGroup()); i++ {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
		}
		b.StartTimer()
		invalidSigners, err := sr.VerifyNodesOnAggSigFailAux()
		b.StopTimer()
		require.Nil(b, err)
		require.NotNil(b, invalidSigners)
	}
}

func BenchmarkSubroundEndRound_VerifyNodesOnAggSigFailTimeParallelThrottle(b *testing.B) {
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer400Sig(container, &statusHandler.AppStatusHandlerMock{})

	signingHandler := &consensusMocks.SigningHandlerStub{
		SignatureShareCalled: func(index uint16) ([]byte, error) {
			return nil, nil
		},
		VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
			return nil
		},
		VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
			return nil
		},
	}
	container.SetSigningHandler(signingHandler)

	sr.Header = &block.Header{}
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		invalidSigners, err := sr.VerifyNodesOnAggSigFailAuxThrottle()
		b.StopTimer()
		require.Nil(b, err)
		require.NotNil(b, invalidSigners)
	}
}
