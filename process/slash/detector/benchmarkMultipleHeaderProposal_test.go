package detector_test

import (
	"testing"

	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/stretchr/testify/require"
)

func BenchmarkMultipleHeaderProposalDetector_ValidateProof(b *testing.B) {
	hasher, err := blake2b.NewBlake2bWithSize(hashSize)
	require.Nil(b, err)

	blsSuite := mcl.NewSuiteBLS12()
	keyGenerator := signing.NewKeyGenerator(blsSuite)
	blsSigners := createMultiSignersBls(metaConsensusGroupSize, hasher, keyGenerator)

	args := createMultipleHeaderDetectorArgs(b, hasher, keyGenerator, blsSigners)
	ssd, err := detector.NewMultipleHeaderProposalsDetector(
		&detector.MultipleHeaderProposalDetectorArgs{
			MultipleHeaderDetectorArgs: args,
		},
	)
	require.NotNil(b, ssd)
	require.Nil(b, err)

	// Worst case scenario: (1/4 * metaConsensusGroupSize + 1) sign the same 3 headers
	noOfMaliciousSigners := uint32(float32(0.25*metaConsensusGroupSize)) + 1
	// For 3 headers, we should have around 8 ms
	// For 3 headers, we should have around 6 ms
	// For 2 headers, we should have around 4 ms
	noOfSignedHeaders := uint64(3)
	slashableHeaders, _ := generateSlashableHeaders(b, hasher, noOfMaliciousSigners, noOfSignedHeaders, args.NodesCoordinator, blsSigners, true)

	slashRes := &coreSlash.SlashingResult{
		SlashingLevel: calcThreatLevel(noOfSignedHeaders),
		Headers:       slashableHeaders,
	}

	proof, err := coreSlash.NewMultipleProposalProof(slashRes)
	require.NotNil(b, proof)
	require.Nil(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = ssd.ValidateProof(proof)
		require.Nil(b, err)
	}
}
