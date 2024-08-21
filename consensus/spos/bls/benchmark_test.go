package bls_test

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto2 "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	multisig2 "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func BenchmarkSubroundSignature_doSignatureJobForManagedKeys63(b *testing.B) {
	benchmarkSubroundSignaturedoSignatureJobForManagedKeys(b, 63)
}

func BenchmarkSubroundSignature_doSignatureJobForManagedKeys400(b *testing.B) {
	benchmarkSubroundSignaturedoSignatureJobForManagedKeys(b, 400)
}

func createMultiSignerSetup(grSize uint16, suite crypto2.Suite) (crypto2.KeyGenerator, map[string]crypto2.PrivateKey) {
	kg := signing.NewKeyGenerator(suite)
	mapKeys := make(map[string]crypto2.PrivateKey)

	for i := uint16(0); i < grSize; i++ {
		sk, pk := kg.GeneratePair()

		pubKey, _ := pk.ToByteArray()
		mapKeys[string(pubKey)] = sk
	}
	return kg, mapKeys
}

func benchmarkSubroundSignaturedoSignatureJobForManagedKeys(b *testing.B, numberOfKeys int) {
	container := mock.InitConsensusCore()
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.EquivalentMessagesFlag
		},
	}
	container.SetEnableEpochsHandler(enableEpochsHandler)
	llSigner := &multisig2.BlsMultiSignerKOSK{}

	suite := mcl.NewSuiteBLS12()
	kg, mapKeys := createMultiSignerSetup(uint16(numberOfKeys), suite)

	multiSigHandler, _ := multisig.NewBLSMultisig(llSigner, kg)

	keysHandlerMock := &testscommon.KeysHandlerStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			return true
		},
		GetHandledPrivateKeyCalled: func(pkBytes []byte) crypto2.PrivateKey {
			return mapKeys[string(pkBytes)]
		},
	}

	args := crypto.ArgsSigningHandler{
		PubKeys: createEligibleListFromMap(mapKeys),
		MultiSignerContainer: &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto2.MultiSigner, error) {
				return multiSigHandler, nil
			}},
		SingleSigner: &cryptoMocks.SingleSignerStub{},
		KeyGenerator: kg,
		KeysHandler:  keysHandlerMock,
	}
	signingHandler, err := crypto.NewSigningHandler(args)
	require.Nil(b, err)

	container.SetSigningHandler(signingHandler)
	consensusState := initConsensusStateWithArgs(keysHandlerMock, numberOfKeys, mapKeys)
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
