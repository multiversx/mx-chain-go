package v2_test

import (
	"sort"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclMultisig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	factoryCrypto "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

const sendProofBenchmarkKeyCount = 400

// generateBLSKeyPairs generates the specified number of BLS key pairs
func generateBLSKeyPairs(kg crypto.KeyGenerator, count int) map[string]crypto.PrivateKey {
	mapKeys := make(map[string]crypto.PrivateKey)

	for i := 0; i < count; i++ {
		sk, pk := kg.GeneratePair()
		pubKey, _ := pk.ToByteArray()
		mapKeys[string(pubKey)] = sk
	}
	return mapKeys
}

// sortedKeysFromMap returns sorted keys from a map of private keys
func sortedKeysFromMap(mapKeys map[string]crypto.PrivateKey) []string {
	keys := make([]string, 0, len(mapKeys))
	for key := range mapKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// BenchmarkSubroundEndRound_SendProof400Keys benchmarks the sendProof method with 400 real BLS keys
func BenchmarkSubroundEndRound_SendProof400Keys(b *testing.B) {
	benchmarkSendProof(b, sendProofBenchmarkKeyCount)
}

func benchmarkSendProof(b *testing.B, numberOfKeys int) {
	container := consensusMocks.InitConsensusCore()

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}
	container.SetEnableEpochsHandler(enableEpochsHandler)

	// Set up real BLS cryptography
	llSigner := &mclMultisig.BlsMultiSignerKOSK{}
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)

	multiSigHandler, err := multisig.NewBLSMultisig(llSigner, kg)
	require.Nil(b, err)

	// Generate real BLS key pairs
	mapKeys := generateBLSKeyPairs(kg, numberOfKeys)
	keys := sortedKeysFromMap(mapKeys)

	keysHandlerMock := &testscommon.KeysHandlerStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			_, exists := mapKeys[string(pkBytes)]
			return exists
		},
		GetHandledPrivateKeyCalled: func(pkBytes []byte) crypto.PrivateKey {
			return mapKeys[string(pkBytes)]
		},
	}

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

	// Set up proofs pool to return false (no proof exists yet)
	container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return false
		},
	})

	// Initialize consensus state with real keys
	consensusState := initializers.InitConsensusStateWithArgsVerifySignature(keysHandlerMock, keys)
	dataToBeSigned := []byte("block_header_hash_for_benchmark")
	consensusState.Data = dataToBeSigned

	ch := make(chan bool, 1)

	sr, err := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		roundTimeDuration,
		0.85,
		0.95,
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	require.Nil(b, err)

	// Set header
	sr.SetHeader(&block.HeaderV2{
		Header: &block.Header{
			Nonce:     1,
			Round:     1,
			TimeStamp: 0,
			ShardID:   0,
		},
	})

	// Set self pub key to be in consensus group (first key)
	sr.SetSelfPubKey(keys[0])

	srEndRound, err := v2.NewSubroundEndRound(
		sr,
		v2.ProcessingThresholdPercent,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{
			ConsensusMetricsCalled: func() spos.ConsensusMetricsHandler {
				consensusMetrics, _ := spos.NewConsensusMetrics(&statusHandler.AppStatusHandlerStub{})
				return consensusMetrics
			},
		},
	)
	require.Nil(b, err)

	// Create signature shares for all validators (simulating that all have signed)
	for i := 0; i < len(keys); i++ {
		_, err := signingHandler.CreateSignatureShareForPublicKey(dataToBeSigned, uint16(i), 0, []byte(keys[i]))
		require.Nil(b, err)
		err = srEndRound.SetJobDone(keys[i], bls.SrSignature, true)
		require.Nil(b, err)
	}

	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		// Reset signature state for each iteration
		for j := 0; j < len(keys); j++ {
			_, _ = signingHandler.CreateSignatureShareForPublicKey(dataToBeSigned, uint16(j), 0, []byte(keys[j]))
		}

		b.StartTimer()
		proofSent, err := srEndRound.SendProof()
		b.StopTimer()

		// The proof should be sent successfully (err may be ErrProofAlreadyPropagated on subsequent runs)
		if err != nil && err != v2.ErrProofAlreadyPropagated && err != v2.ErrTimeOut {
			require.Nil(b, err)
		}
		_ = proofSent
	}
}

// BenchmarkSubroundEndRound_SendProof63Keys benchmarks the sendProof method with 63 keys (standard consensus size)
func BenchmarkSubroundEndRound_SendProof63Keys(b *testing.B) {
	benchmarkSendProof(b, 63)
}

// BenchmarkSubroundEndRound_SendProof100Keys benchmarks the sendProof method with 100 keys
func BenchmarkSubroundEndRound_SendProof100Keys(b *testing.B) {
	benchmarkSendProof(b, 100)
}

// BenchmarkSubroundEndRound_SendProof200Keys benchmarks the sendProof method with 200 keys
func BenchmarkSubroundEndRound_SendProof200Keys(b *testing.B) {
	benchmarkSendProof(b, 200)
}
