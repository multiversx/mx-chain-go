package bls_test

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func BenchmarkSubroundSignature_doSignatureJobForManagedKeys63(b *testing.B) {
	benchmarkSubroundSignaturedoSignatureJobForManagedKeys(b, 63)
}

func BenchmarkSubroundSignature_doSignatureJobForManagedKeys400(b *testing.B) {
	benchmarkSubroundSignaturedoSignatureJobForManagedKeys(b, 400)
}

func benchmarkSubroundSignaturedoSignatureJobForManagedKeys(b *testing.B, numberOfKeys int) {
	container := mock.InitConsensusCore()
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.EquivalentMessagesFlag
		},
	}
	container.SetEnableEpochsHandler(enableEpochsHandler)

	signingHandler := &consensusMocks.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			return []byte("SIG"), nil
		},
	}
	container.SetSigningHandler(signingHandler)
	consensusState := initConsensusStateWithKeysHandlerWithGroupSize(&testscommon.KeysHandlerStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			return true
		},
	},
		numberOfKeys,
	)
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	signatureSentForPks := make(map[string]struct{})
	mutex := sync.Mutex{}
	srSignature, _ := bls.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{
			SignatureSentCalled: func(pkBytes []byte) {
				mutex.Lock()
				signatureSentForPks[string(pkBytes)] = struct{}{}
				mutex.Unlock()
			},
		},
		&mock.SposWorkerMock{},
	)

	sr.Header = &block.Header{}
	signaturesBroadcast := make(map[string]int)
	container.SetBroadcastMessenger(&mock.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			mutex.Lock()
			signaturesBroadcast[string(message.PubKey)]++
			mutex.Unlock()
			return nil
		},
	})

	sr.SetSelfPubKey("OTHER")

	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		r := srSignature.DoSignatureJobForManagedKeys()
		b.StopTimer()

		require.True(b, r)
	}

}
