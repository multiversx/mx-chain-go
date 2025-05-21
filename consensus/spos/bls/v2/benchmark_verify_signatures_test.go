package v2_test

import (
	"context"
	"sort"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/stretchr/testify/require"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	mclMultisig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"

	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"

	"github.com/multiversx/mx-chain-go/common"
	factoryCrypto "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

const benchmarkKeyPairsCardinal = 400

// createListFromMapKeys make a predictable iteration on keys from a map of keys
func createListFromMapKeys(mapKeys map[string]crypto.PrivateKey) []string {
	keys := make([]string, 0, len(mapKeys))

	for key := range mapKeys {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// generateKeyPairs generates benchmarkKeyPairsCardinal number of pairs(public key & private key)
func generateKeyPairs(kg crypto.KeyGenerator) map[string]crypto.PrivateKey {
	mapKeys := make(map[string]crypto.PrivateKey)

	for i := uint16(0); i < benchmarkKeyPairsCardinal; i++ {
		sk, pk := kg.GeneratePair()

		pubKey, _ := pk.ToByteArray()
		mapKeys[string(pubKey)] = sk
	}
	return mapKeys
}

// BenchmarkSubroundEndRound_VerifyNodesOnAggSigFailTime measure time needed to verify signatures
func BenchmarkSubroundEndRound_VerifyNodesOnAggSigFailTime(b *testing.B) {

	b.ResetTimer()
	b.StopTimer()
	ctx, cancel := context.WithCancel(context.TODO())

	defer func() {
		cancel()
	}()

	container := consensus.InitConsensusCore()
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}
	container.SetEnableEpochsHandler(enableEpochsHandler)
	llSigner := &mclMultisig.BlsMultiSignerKOSK{}
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)

	multiSigHandler, _ := multisig.NewBLSMultisig(llSigner, kg)

	mapKeys := generateKeyPairs(kg)

	keysHandlerMock := &testscommon.KeysHandlerStub{
		GetHandledPrivateKeyCalled: func(pkBytes []byte) crypto.PrivateKey {
			return mapKeys[string(pkBytes)]
		},
	}
	keys := createListFromMapKeys(mapKeys)
	args := factoryCrypto.ArgsSigningHandler{
		PubKeys: keys,
		MultiSignerContainer: &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return multiSigHandler, nil
			},
		},
		SingleSigner: &cryptoMocks.SingleSignerStub{},
		KeyGenerator: kg,
		KeysHandler:  keysHandlerMock,
	}

	signingHandler, err := factoryCrypto.NewSigningHandler(args)
	require.Nil(b, err)

	container.SetSigningHandler(signingHandler)
	consensusState := initializers.InitConsensusStateWithArgsVerifySignature(keysHandlerMock, keys)
	dataToBeSigned := []byte("message")
	consensusState.Data = dataToBeSigned

	sr := initSubroundEndRoundWithContainerAndConsensusState(container, &statusHandler.AppStatusHandlerStub{}, consensusState, &dataRetrieverMocks.ThrottlerStub{})
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
}
