package multisig

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	llsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	"github.com/stretchr/testify/assert"
)

func createKeysAndMultiSignerBlsKOSK(
	grSize uint16,
	suite crypto.Suite,
) ([][]byte, [][]byte, crypto.MultiSigner) {

	kg, privKeys, pubKeys := createMultiSignerSetup(grSize, suite)
	llSigner := &llsig.BlsMultiSignerKOSK{}
	multiSigner, _ := multisig.NewBLSMultisig(llSigner, kg)

	return privKeys, pubKeys, multiSigner
}

func createKeysAndMultiSignerBlsNoKOSK(
	grSize uint16,
	hasher hashing.Hasher,
	suite crypto.Suite,
) ([][]byte, [][]byte, crypto.MultiSigner) {
	kg, privKeys, pubKeys := createMultiSignerSetup(grSize, suite)
	llSigner := &llsig.BlsMultiSigner{Hasher: hasher}

	multiSigner, _ := multisig.NewBLSMultisig(llSigner, kg)

	return privKeys, pubKeys, multiSigner
}

func createMultiSignerSetup(grSize uint16, suite crypto.Suite) (crypto.KeyGenerator, [][]byte, [][]byte) {
	kg := signing.NewKeyGenerator(suite)
	privKeys := make([][]byte, grSize)
	pubKeys := make([][]byte, grSize)

	for i := uint16(0); i < grSize; i++ {
		sk, pk := kg.GeneratePair()
		privKeys[i], _ = sk.ToByteArray()
		pubKeys[i], _ = pk.ToByteArray()
	}
	return kg, privKeys, pubKeys
}

func createSignaturesShares(privKeys [][]byte, multiSigner crypto.MultiSigner, message []byte) [][]byte {
	sigShares := make([][]byte, len(privKeys))
	for i := uint16(0); i < uint16(len(privKeys)); i++ {
		sigShares[i], _ = multiSigner.CreateSignatureShare(privKeys[i], message)
	}

	return sigShares
}

func TestMultiSig_Bls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Parallel()

	numSigners := uint16(6)
	message := "message"
	hashSize := 16
	hasher, _ := blake2b.NewBlake2bWithSize(hashSize)
	suite := mcl.NewSuiteBLS12()

	privKeys, pubKeys, multiSigner := createKeysAndMultiSignerBlsNoKOSK(numSigners, hasher, suite)

	numOfTimesToRepeatTests := 100
	for currentIdx := 0; currentIdx < numOfTimesToRepeatTests; currentIdx++ {
		message = fmt.Sprintf("%s%d", message, currentIdx)
		signatures := createSignaturesShares(privKeys, multiSigner, []byte(message))

		aggSig, err := multiSigner.AggregateSigs(pubKeys, signatures)
		assert.Nil(t, err)
		assert.NotNil(t, aggSig)

		err = multiSigner.VerifyAggregatedSig(pubKeys, []byte(message), aggSig)
		assert.Nil(t, err)
	}
}
