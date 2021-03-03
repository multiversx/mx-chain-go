package multisig

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	llsig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/stretchr/testify/assert"
)

func createMultiSignersBls(
	numOfSigners uint16,
	grSize uint16,
	hasher hashing.Hasher,
	suite crypto.Suite,
) ([]string, []crypto.MultiSigner) {

	kg := signing.NewKeyGenerator(suite)

	var pubKeyBytes []byte

	privKeys := make([]crypto.PrivateKey, grSize)
	pubKeys := make([]crypto.PublicKey, grSize)
	pubKeysStr := make([]string, grSize)

	for i := uint16(0); i < grSize; i++ {
		sk, pk := kg.GeneratePair()
		privKeys[i] = sk
		pubKeys[i] = pk

		pubKeyBytes, _ = pk.ToByteArray()
		pubKeysStr[i] = string(pubKeyBytes)
	}

	multiSigners := make([]crypto.MultiSigner, numOfSigners)
	llSigner := &llsig.BlsMultiSigner{Hasher: hasher}

	for i := uint16(0); i < numOfSigners; i++ {
		multiSigners[i], _ = multisig.NewBLSMultisig(llSigner, pubKeysStr, privKeys[i], kg, i)
	}

	return pubKeysStr, multiSigners
}

func createSignaturesShares(numOfSigners uint16, multiSigners []crypto.MultiSigner, message []byte) [][]byte {
	sigShares := make([][]byte, numOfSigners)
	for i := uint16(0); i < numOfSigners; i++ {
		sigShares[i], _ = multiSigners[i].CreateSignatureShare(message, []byte(""))
	}

	return sigShares
}

func setSignatureSharesAllSignersBls(multiSigners []crypto.MultiSigner, sigs [][]byte) error {
	grSize := uint16(len(multiSigners))
	var err error

	for i := uint16(0); i < grSize; i++ {
		for j := uint16(0); j < grSize; j++ {
			err = multiSigners[j].StoreSignatureShare(i, sigs[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func verifySigAllSignersBls(
	multiSigners []crypto.MultiSigner,
	message []byte,
	signature []byte,
	pubKeys []string,
	bitmap []byte,
	grSize uint16,
) error {

	var err error
	var muSig crypto.MultiSigner

	for i := uint16(0); i < grSize; i++ {
		muSig, err = multiSigners[i].Create(pubKeys, i)
		if err != nil {
			return err
		}

		muSigBls, ok := muSig.(crypto.MultiSigner)
		if !ok {
			return crypto.ErrInvalidSigner
		}

		multiSigners[i] = muSigBls
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

func aggregateSignatureSharesAllSignersBls(multiSigners []crypto.MultiSigner, bitmap []byte, grSize uint16) (
	signature []byte,
	err error,
) {
	aggSig, err := multiSigners[0].AggregateSigs(bitmap)

	if err != nil {
		return nil, err
	}

	for i := uint16(1); i < grSize; i++ {
		aggSig2, err1 := multiSigners[i].AggregateSigs(bitmap)

		if err1 != nil {
			return nil, err1
		}

		if !bytes.Equal(aggSig, aggSig2) {
			return nil, errors.New("aggregated signatures not equal")
		}
	}

	return aggSig, nil
}

func TestMultiSig_Bls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Parallel()

	consensusGroupSize := uint16(6)
	numOfSigners := uint16(6)
	message := "message"

	bitmapSize := consensusGroupSize/8 + 1
	// set bitmap to select all members
	bitmap := make([]byte, bitmapSize)
	byteMask := 0xFF
	for i := uint16(0); i < bitmapSize; i++ {
		bitmap[i] = byte((((1 << consensusGroupSize) - 1) >> i) & byteMask)
	}

	hashSize := 16
	hasher, _ := blake2b.NewBlake2bWithSize(hashSize)
	suite := mcl.NewSuiteBLS12()

	pubKeysStr, multiSigners := createMultiSignersBls(numOfSigners, consensusGroupSize, hasher, suite)

	numOfTimesToRepeatTests := 100
	for currentIdx := 0; currentIdx < numOfTimesToRepeatTests; currentIdx++ {
		message = fmt.Sprintf("%s%d", message, currentIdx)
		signatures := createSignaturesShares(numOfSigners, multiSigners, []byte(message))

		err := setSignatureSharesAllSignersBls(multiSigners, signatures)
		assert.Nil(t, err)

		aggSig, err := aggregateSignatureSharesAllSignersBls(multiSigners, bitmap, consensusGroupSize)
		assert.Nil(t, err)
		assert.NotNil(t, aggSig)

		err = verifySigAllSignersBls(multiSigners, []byte(message), aggSig, pubKeysStr, bitmap, consensusGroupSize)
		assert.Nil(t, err)
	}
}
