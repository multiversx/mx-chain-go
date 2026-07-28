package components

import (
	"testing"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/stretchr/testify/require"
)

func TestFastConsensusSigner_BindsSharesAndAggregateToKeysAndMessage(t *testing.T) {
	keyGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	firstPrivateKey, firstPublicKey := keyGenerator.GeneratePair()
	secondPrivateKey, secondPublicKey := keyGenerator.GeneratePair()

	firstSigner, err := newFastConsensusSigner(firstPrivateKey, firstPublicKey, keyGenerator)
	require.NoError(t, err)
	secondSigner, err := newFastConsensusSigner(secondPrivateKey, secondPublicKey, keyGenerator)
	require.NoError(t, err)

	message := []byte("consensus payload")
	firstShare, err := firstSigner.CreateSignatureShareV2(firstPrivateKey, message)
	require.NoError(t, err)
	secondShare, err := secondSigner.CreateSignatureShareV2(secondPrivateKey, message)
	require.NoError(t, err)
	require.Len(t, firstShare, fastConsensusSignatureSize)
	require.NotEqual(t, firstShare, secondShare)

	repeatedShare, err := firstSigner.CreateSignatureShareV2(firstPrivateKey, message)
	require.NoError(t, err)
	require.Equal(t, firstShare, repeatedShare)

	require.NoError(t, firstSigner.VerifySignatureShareV2(firstPublicKey, message, firstShare))
	require.Error(t, firstSigner.VerifySignatureShareV2(firstPublicKey, []byte("other payload"), firstShare))
	require.Error(t, firstSigner.VerifySignatureShareV2(secondPublicKey, message, firstShare))

	aggregatedSignature, err := firstSigner.AggregateSigsV2(
		[]crypto.PublicKey{firstPublicKey, secondPublicKey},
		[][]byte{firstShare, secondShare},
	)
	require.NoError(t, err)
	require.Len(t, aggregatedSignature, fastConsensusSignatureSize)
	require.NoError(t, firstSigner.VerifyAggregatedSigV2(
		[]crypto.PublicKey{firstPublicKey, secondPublicKey},
		message,
		aggregatedSignature,
	))

	repeatedAggregate, err := firstSigner.AggregateSigsV2(
		[]crypto.PublicKey{firstPublicKey, secondPublicKey},
		[][]byte{firstShare, secondShare},
	)
	require.NoError(t, err)
	require.Equal(t, aggregatedSignature, repeatedAggregate)

	require.Error(t, firstSigner.VerifyAggregatedSigV2(
		[]crypto.PublicKey{secondPublicKey, firstPublicKey},
		message,
		aggregatedSignature,
	))
	require.Error(t, firstSigner.VerifyAggregatedSigV2(
		[]crypto.PublicKey{firstPublicKey, secondPublicKey},
		[]byte("other payload"),
		aggregatedSignature,
	))
	require.Error(t, firstSigner.VerifyAggregatedSigV2(
		[]crypto.PublicKey{firstPublicKey},
		message,
		aggregatedSignature,
	))

	tamperedShare := append([]byte(nil), firstShare...)
	tamperedShare[0] ^= 1
	tamperedAggregate, err := firstSigner.AggregateSigsV2(
		[]crypto.PublicKey{firstPublicKey, secondPublicKey},
		[][]byte{tamperedShare, secondShare},
	)
	require.NoError(t, err)
	require.Error(t, firstSigner.VerifyAggregatedSigV2(
		[]crypto.PublicKey{firstPublicKey, secondPublicKey},
		message,
		tamperedAggregate,
	))
}

func TestFastConsensusSigner_RejectsMalformedSharesDuringAggregation(t *testing.T) {
	keyGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	privateKey, publicKey := keyGenerator.GeneratePair()
	signer, err := newFastConsensusSigner(privateKey, publicKey, keyGenerator)
	require.NoError(t, err)

	_, err = signer.AggregateSigsV2(
		[]crypto.PublicKey{publicKey},
		[][]byte{[]byte("not a signature")},
	)
	require.ErrorIs(t, err, errInvalidFastConsensusSignature)
}
