package detector_test

import (
	"testing"

	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/stretchr/testify/require"
)

func BenchmarkMultipleHeaderProposalDetector_VerifyData(b *testing.B) {
	hasher, err := blake2b.NewBlake2bWithSize(hashSize)
	require.Nil(b, err)

	blsSuite := mcl.NewSuiteBLS12()
	keyGenerator := signing.NewKeyGenerator(blsSuite)
	blsSigners := createMultiSignersBls(metaConsensusGroupSize, hasher, keyGenerator)

	noOfSignedHeaders := uint32(3)
	args := createMultipleHeaderDetectorArgs(b, hasher, keyGenerator, blsSigners)
	slashableHeaders, _ := generateSlashableHeaders(b, hasher, maxNoOfMaliciousValidatorsOnMetaChain, noOfSignedHeaders, args.NodesCoordinator, blsSigners, true)
	interceptedHeaders := createInterceptedHeaders(slashableHeaders)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.Run("", func(b *testing.B) {
			benchmarkVerifyDataMultipleHeaderProposal(b, hasher, keyGenerator, blsSigners, interceptedHeaders)
		})
	}
}

func benchmarkVerifyDataMultipleHeaderProposal(
	b *testing.B,
	hasher hashing.Hasher,
	keyGenerator crypto.KeyGenerator,
	blsSigners map[string]multiSignerData,
	interceptedHeaders []process.InterceptedHeader,
) {
	args := createMultipleHeaderDetectorArgs(b, hasher, keyGenerator, blsSigners)
	ssd, err := detector.NewMultipleHeaderProposalsDetector(args)
	require.Nil(b, err)

	b.ResetTimer()
	verifyInterceptedHeaders(b, interceptedHeaders, ssd)
}

func BenchmarkMultipleHeaderProposalDetector_ValidateProof(b *testing.B) {
	hasher, err := blake2b.NewBlake2bWithSize(hashSize)
	require.Nil(b, err)

	blsSuite := mcl.NewSuiteBLS12()
	keyGenerator := signing.NewKeyGenerator(blsSuite)
	blsSigners := createMultiSignersBls(metaConsensusGroupSize, hasher, keyGenerator)

	args := createMultipleHeaderDetectorArgs(b, hasher, keyGenerator, blsSigners)
	ssd, err := detector.NewMultipleHeaderProposalsDetector(args)
	require.Nil(b, err)

	// For 4 headers, it should take around 8 ms to verify the proof
	// For 3 headers, it should take around 6 ms to verify the proof
	// For 2 headers, it should take around 4 ms to verify the proof
	noOfSlashableHeaders := uint32(3)
	slashRes := generateMultipleHeaderProposalSlashResult(b, hasher, maxNoOfMaliciousValidatorsOnMetaChain, noOfSlashableHeaders, args.NodesCoordinator, blsSigners)

	proof, err := coreSlash.NewMultipleProposalProof(slashRes)
	require.NotNil(b, proof)
	require.Nil(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = ssd.ValidateProof(proof)
		require.Nil(b, err)
	}
}
