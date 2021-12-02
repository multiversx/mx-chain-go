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

func BenchmarkMultipleHeaderSigningDetector_VerifyData(b *testing.B) {
	hasher, err := blake2b.NewBlake2bWithSize(hashSize)
	require.Nil(b, err)

	blsSuite := mcl.NewSuiteBLS12()
	keyGenerator := signing.NewKeyGenerator(blsSuite)
	blsSigners := createMultiSignersBls(metaConsensusGroupSize, hasher, keyGenerator)

	args := createHeaderSigningDetectorArgs(b, hasher, keyGenerator, blsSigners)
	// Worst case scenario: same (1/4 * metaConsensusGroupSize + 1) validators sign 3 different headers
	noOfSignedHeaders := uint32(3)
	slashableHeaders, _ := generateSlashableHeaders(b, hasher, maxNoOfMaliciousValidatorsOnMetaChain, noOfSignedHeaders, args.NodesCoordinator, blsSigners, false)
	interceptedHeaders := createInterceptedHeaders(slashableHeaders)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.Run("", func(b *testing.B) {
			benchmarkVerifyDataMultipleHeaderSigning(b, hasher, keyGenerator, blsSigners, interceptedHeaders)
		})
	}
}

func benchmarkVerifyDataMultipleHeaderSigning(
	b *testing.B,
	hasher hashing.Hasher,
	keyGenerator crypto.KeyGenerator,
	blsSigners map[string]multiSignerData,
	interceptedHeaders []process.InterceptedHeader) {
	args := createHeaderSigningDetectorArgs(b, hasher, keyGenerator, blsSigners)
	ssd, err := detector.NewMultipleHeaderSigningDetector(args)
	require.Nil(b, err)

	b.ResetTimer()
	verifyInterceptedHeaders(b, interceptedHeaders, ssd)
}

func BenchmarkMultipleHeaderSigningDetector_ValidateProof(b *testing.B) {
	hasher, err := blake2b.NewBlake2bWithSize(hashSize)
	require.Nil(b, err)

	blsSuite := mcl.NewSuiteBLS12()
	keyGenerator := signing.NewKeyGenerator(blsSuite)
	blsSigners := createMultiSignersBls(metaConsensusGroupSize, hasher, keyGenerator)

	args := createHeaderSigningDetectorArgs(b, hasher, keyGenerator, blsSigners)
	ssd, err := detector.NewMultipleHeaderSigningDetector(args)
	require.Nil(b, err)

	// Worst case scenario: same (1/4 * metaConsensusGroupSize + 1) validators sign 3 different headers
	noOfSignedHeaders := uint32(3)
	slashRes := generateMultipleHeaderSigningSlashResult(b, hasher, maxNoOfMaliciousValidatorsOnMetaChain, noOfSignedHeaders, args.NodesCoordinator, blsSigners)
	proof, err := coreSlash.NewMultipleSigningProof(slashRes)
	require.NotNil(b, proof)
	require.Nil(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = ssd.ValidateProof(proof)
		require.Nil(b, err)
	}
}

func createHeaderSigningDetectorArgs(
	b *testing.B,
	hasher hashing.Hasher,
	keyGenerator crypto.KeyGenerator,
	multiSignersData map[string]multiSignerData,
) *detector.MultipleHeaderSigningDetectorArgs {
	detectorArgs := createMultipleHeaderDetectorArgs(b, hasher, keyGenerator, multiSignersData)

	return &detector.MultipleHeaderSigningDetectorArgs{
		MultipleHeaderDetectorArgs: detectorArgs,
		RoundHashCache:             detector.NewRoundHashCache(cacheSize),
	}
}
