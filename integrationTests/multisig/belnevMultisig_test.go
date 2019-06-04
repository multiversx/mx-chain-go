package multisig

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/stretchr/testify/assert"
)

type multiSignerBN interface {
	crypto.MultiSigner
	// CreateCommitment creates a secret commitment and the corresponding public commitment point
	CreateCommitment() (commSecret []byte, commitment []byte)
	// StoreCommitmentHash adds a commitment hash to the list with the specified position
	StoreCommitmentHash(index uint16, commHash []byte) error
	// CommitmentHash returns the commitment hash from the list with the specified position
	CommitmentHash(index uint16) ([]byte, error)
	// StoreCommitment adds a commitment to the list with the specified position
	StoreCommitment(index uint16, value []byte) error
	// Commitment returns the commitment from the list with the specified position
	Commitment(index uint16) ([]byte, error)
	// AggregateCommitments aggregates the list of commitments
	AggregateCommitments(bitmap []byte) error
}

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
) (mutiSigners []multiSignerBN, err error) {

	groupSize := uint16(len(pubKeysStr))
	multiSigners := make([]multiSignerBN, groupSize)

	for i := uint16(0); i < groupSize; i++ {
		multiSigners[i], err = multisig.NewBelNevMultisig(hasher, pubKeysStr, privKeys[i], kg, i)
		if err != nil {
			return nil, err
		}
	}

	return multiSigners, nil
}

func createAndSetCommitment(multiSig multiSignerBN) (commSecret, comm []byte) {
	commSecret, comm = multiSig.CreateCommitment()

	return commSecret, comm
}

func createAndSetCommitmentsAllSigners(multiSigners []multiSignerBN) error {
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

func aggregateCommitmentsForAllSigners(multiSigners []multiSignerBN, bitmap []byte, grSize uint16) error {
	for i := uint16(0); i < grSize; i++ {
		err := multiSigners[i].AggregateCommitments(bitmap)

		if err != nil {
			return err
		}
	}

	return nil
}

func createAndSetSignatureSharesAllSigners(multiSigners []multiSignerBN, msg []byte, bitmap []byte) error {
	grSize := uint16(len(multiSigners))
	sigShares := make([][]byte, grSize)
	var err error

	for i := uint16(0); i < grSize; i++ {
		sigShares[i], err = multiSigners[i].CreateSignatureShare(msg, bitmap)
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

func aggregateSignatureSharesAllSigners(multiSigners []multiSignerBN, bitmap []byte, grSize uint16) (
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
	multiSigners []multiSignerBN,
	message []byte,
	signature []byte,
	pubKeys []string,
	bitmap []byte,
	grSize uint16) error {

	var err error
	var muSig crypto.MultiSigner

	for i := uint16(0); i < grSize; i++ {
		muSig, err = multiSigners[i].Create(pubKeys, i)
		if err != nil {
			return err
		}

		muSigBn, ok := muSig.(multiSignerBN)
		if !ok {
			return crypto.ErrInvalidSigner
		}

		multiSigners[i] = muSigBn
		err = multiSigners[i].SetAggregatedSig(signature)
		if err != nil {
			return err
		}

		err = multiSigners[i].Verify(message, bitmap)
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
	assert.Nil(t, err)

	err = createAndSetSignatureSharesAllSigners(multiSigners, message, bitmap)
	assert.Nil(t, err)

	aggSig, err := aggregateSignatureSharesAllSigners(multiSigners, bitmap, consensusGroupSize)
	assert.Nil(t, err)

	err = verifySigAllSigners(multiSigners, message, aggSig, pubKeysStr, bitmap, consensusGroupSize)
	assert.Nil(t, err)
}
