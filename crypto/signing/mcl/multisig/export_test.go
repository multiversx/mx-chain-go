package multisig

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/herumi/bls-go-binary/bls"
)

func (bms *BlsMultiSigner) ScalarMulSig(suite crypto.Suite, scalarBytes []byte, sigPoint *mcl.PointG1) (*mcl.PointG1, error) {
	return bms.scalarMulSig(suite, scalarBytes, sigPoint)
}

func PreparePublicKeys(pubKeys []crypto.PublicKey, hasher hashing.Hasher, suite crypto.Suite) ([]bls.PublicKey, error) {
	return preparePublicKeys(pubKeys, hasher, suite)
}

func (bms *BlsMultiSigner) PrepareSignatures(suite crypto.Suite, signatures [][]byte, pubKeysSigners []crypto.PublicKey) ([]bls.Sign, error) {
	return bms.prepareSignatures(suite, signatures, pubKeysSigners)
}

func ScalarMulPk(suite crypto.Suite, scalarBytes []byte, pk crypto.Point) (crypto.Point, error) {
	return scalarMulPk(suite, scalarBytes, pk)
}

func HashPublicKeyPoints(hasher hashing.Hasher, pubKeyPoint crypto.Point, concatPubKeys []byte) ([]byte, error) {
	return hashPublicKeyPoints(hasher, pubKeyPoint, concatPubKeys)
}

func ConcatPubKeys(pubKeys []crypto.PublicKey) ([]byte, error) {
	return concatPubKeys(pubKeys)
}
