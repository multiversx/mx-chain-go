package bls_test

import (
	"context"
	"sort"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/stretchr/testify/require"

	crypto2 "github.com/multiversx/mx-chain-crypto-go"
	multisig2 "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"

	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/dataRetriever/mock"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

// Function to make a predictable iteration on keys from a map of keys
func createListFromMapKeys(mapKeys map[string]crypto2.PrivateKey) []string {
	keys := make([]string, 0, len(mapKeys))

	for key := range mapKeys {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// Benchmark on measuring performance
func BenchmarkSubroundEndRound_VerifyNodesOnAggSigFailTime(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()
	ctx, cancel := context.WithCancel(context.TODO())
	container := mock.InitConsensusCore()
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.EquivalentMessagesFlag
		},
	}
	container.SetEnableEpochsHandler(enableEpochsHandler)
	llSigner := &multisig2.BlsMultiSignerKOSK{}
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	mapKeys := make(map[string]crypto2.PrivateKey)

	for i := uint16(0); i < 400; i++ {
		sk, pk := kg.GeneratePair()

		pubKey, _ := pk.ToByteArray()
		mapKeys[string(pubKey)] = sk
	}

	multiSigHandler, _ := multisig.NewBLSMultisig(llSigner, kg)

	keysHandlerMock := &testscommon.KeysHandlerStub{
		GetHandledPrivateKeyCalled: func(pkBytes []byte) crypto2.PrivateKey {
			return mapKeys[string(pkBytes)]
		},
	}
	keys := createListFromMapKeys(mapKeys)
	args := crypto.ArgsSigningHandler{
		PubKeys: keys,
		MultiSignerContainer: &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto2.MultiSigner, error) {
				return multiSigHandler, nil
			},
		},
		SingleSigner: &cryptoMocks.SingleSignerStub{},
		KeyGenerator: kg,
		KeysHandler:  keysHandlerMock,
	}

	signingHandler, err := crypto.NewSigningHandler(args)
	require.Nil(b, err)

	container.SetSigningHandler(signingHandler)
	consensusState := initConsensusStateWithArgsVerifySignature(keysHandlerMock, keys)
	dataToBeSigned := []byte("message")
	consensusState.Data = dataToBeSigned

	sr := *initSubroundEndRoundWithContainer400Sig(container, &statusHandler.AppStatusHandlerStub{}, consensusState, &dataRetrieverMocks.ThrottlerStub{})
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		_, err := sr.SigningHandler().CreateSignatureShareForPublicKey(dataToBeSigned, uint16(i), sr.EnableEpochsHandler().GetCurrentEpoch(), []byte(keys[i]))
		require.Nil(b, err)
		_ = sr.SetJobDone(keys[i], bls.SrSignature, true)
	}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		invalidSigners, err := sr.VerifyNodesOnAggSigFail(ctx)
		b.StopTimer()
		require.Nil(b, err)
		require.NotNil(b, invalidSigners)
	}
	cancel()
}
