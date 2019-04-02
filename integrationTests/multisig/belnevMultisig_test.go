package multisig

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func generateKeyPairs(kg crypto.KeyGenerator, consensusSize uint16) (
	privKeys []crypto.PrivateKey,
	pubKeysStr []string) {

	privKeys = make([]crypto.PrivateKey, consensusSize)
	pubKeys := make([]crypto.PublicKey, consensusSize)
	pubKeysStr = make([]string, consensusSize)

	for i := uint16(0); i < consensusSize; i++ {
		sk, pk := kg.GeneratePair()
		privKeys[i] = sk
		pubKeys[i] = pk

		pubKeyBytes, _ := pk.ToByteArray()
		pubKeysStr[i] = string(pubKeyBytes)
	}

	return privKeys, pubKeysStr
}

func createMultiSigners(
	kg crypto.KeyGenerator,
	hasher hashing.Hasher,
	privKeys []crypto.PrivateKey,
	pubKeysStr []string,
) (mutiSigners []crypto.MultiSigner, err error) {

	groupSize := uint16(len(pubKeysStr))
	multiSigners := make([]crypto.MultiSigner, groupSize)

	for i := uint16(0); i < groupSize; i++ {
		multiSigners[i], err = multisig.NewBelNevMultisig(hasher, pubKeysStr, privKeys[i], kg, i)
		if err != nil {
			return nil, err
		}
	}

	return multiSigners, nil
}

func createAndSetCommitment(multiSig crypto.MultiSigner) (commSecret, comm []byte) {
	commSecret, comm = multiSig.CreateCommitment()

	return commSecret, comm
}

func createAndSetCommitmentsAllSigners(multiSigners []crypto.MultiSigner) error {
	groupSize := uint16(len(multiSigners))
	var err error

	for i := uint16(0); i < groupSize; i++ {
		_, comm := createAndSetCommitment(multiSigners[i])

		for j := uint16(0); j < groupSize; j++ {
			err = multiSigners[j].StoreCommitment(i, comm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func aggregateCommitmentsForAllSigners(multiSigners []crypto.MultiSigner, bitmap []byte, grSize uint16) error {
	for i := uint16(0); i < grSize; i++ {
		err := multiSigners[i].AggregateCommitments(bitmap)

		if err != nil {
			return err
		}
	}

	return nil
}

func setMessageAllSigners(multiSigners []crypto.MultiSigner, msg []byte) error {
	grSize := uint16(len(multiSigners))

	for i := uint16(0); i < grSize; i++ {
		err := multiSigners[i].SetMessage(msg)

		if err != nil {
			return err
		}
	}

	return nil
}

func createAndSetSignatureSharesAllSigners(multiSigners []crypto.MultiSigner, bitmap []byte) error {
	grSize := uint16(len(multiSigners))
	sigShares := make([][]byte, grSize)
	var err error

	for i := uint16(0); i < grSize; i++ {
		sigShares[i], err = multiSigners[i].CreateSignatureShare(bitmap)
		if err != nil {
			return err
		}

		for j := uint16(0); j < grSize; j++ {
			err = multiSigners[j].StoreSignatureShare(i, sigShares[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func aggregateSignatureSharesAllSigners(multiSigners []crypto.MultiSigner, bitmap []byte, grSize uint16) (
	signature []byte,
	err error,
) {
	aggSig, err := multiSigners[0].AggregateSigs(bitmap)

	if err != nil {
		return nil, err
	}

	for i := uint16(1); i < grSize; i++ {
		aggSig2, err := multiSigners[i].AggregateSigs(bitmap)

		if err != nil {
			return nil, err
		}

		if !bytes.Equal(aggSig, aggSig2) {
			return nil, errors.New("aggregated signatures not equal")
		}
	}

	return aggSig, nil
}

func verifySigAllSigners(
	multiSigners []crypto.MultiSigner,
	message []byte,
	signature []byte,
	pubKeys []string,
	bitmap []byte,
	grSize uint16) error {

	var err error

	for i := uint16(0); i < grSize; i++ {
		multiSigners[i], err = multiSigners[i].Create(pubKeys, i)
		if err != nil {
			return err
		}

		err = multiSigners[i].SetMessage(message)
		if err != nil {
			return err
		}

		err = multiSigners[i].SetAggregatedSig(signature)
		if err != nil {
			return err
		}

		err = multiSigners[i].Verify(bitmap)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestBelnev_MultiSigningMultipleSignersOK(t *testing.T) {
	consensusGroupSize := uint16(21)
	suite := kyber.NewBlakeSHA256Ed25519()
	kg := signing.NewKeyGenerator(suite)

	privKeys, pubKeysStr := generateKeyPairs(kg, consensusGroupSize)
	hasher := sha256.Sha256{}

	multiSigners, err := createMultiSigners(kg, hasher, privKeys, pubKeysStr)
	assert.Nil(t, err)

	err = createAndSetCommitmentsAllSigners(multiSigners)
	assert.Nil(t, err)

	bitmapSize := consensusGroupSize/8 + 1
	// set bitmap to select all 21 members
	bitmap := make([]byte, bitmapSize)
	byteMask := 0xFF

	for i := uint16(0); i < bitmapSize; i++ {
		bitmap[i] = byte((((1 << consensusGroupSize) - 1) >> i) & byteMask)
	}

	err = aggregateCommitmentsForAllSigners(multiSigners, bitmap, consensusGroupSize)
	assert.Nil(t, err)

	message := []byte("message to be signed")
	err = setMessageAllSigners(multiSigners, message)
	assert.Nil(t, err)

	err = createAndSetSignatureSharesAllSigners(multiSigners, bitmap)
	assert.Nil(t, err)

	aggSig, err := aggregateSignatureSharesAllSigners(multiSigners, bitmap, consensusGroupSize)
	assert.Nil(t, err)

	err = verifySigAllSigners(multiSigners, message, aggSig, pubKeysStr, bitmap, consensusGroupSize)
	assert.Nil(t, err)
}
