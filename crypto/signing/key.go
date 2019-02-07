package signing

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

// privateKey holds the private key and the chosen curve
type privateKey struct {
	suite crypto.Suite
	sk    crypto.Scalar
}

// publicKey holds the public key and the chosen curve
type publicKey struct {
	suite crypto.Suite
	pk    crypto.Point
}

// Sign creates a signature of the message using the current private key
func (spk *privateKey) Sign(message []byte, signer crypto.SingleSigner) ([]byte, error) {
	if signer == nil {
		return nil, crypto.ErrNilSingleSigner
	}

	return signer.Sign(spk.suite, spk.sk, message)
}

// ToByteArray returns the byte array representation of the private key
func (spk *privateKey) ToByteArray() ([]byte, error) {
	return spk.sk.MarshalBinary()
}

// GeneratePublic builds a public key for the current private key
func (spk *privateKey) GeneratePublic() crypto.PublicKey {
	point := spk.suite.CreatePoint().Base()
	privKey, _ := point.Mul(spk.sk)
	return &publicKey{
		suite: spk.suite,
		pk:    privKey,
	}
}

// Suite returns the Suite (curve data) used for this private key
func (spk *privateKey) Suite() crypto.Suite {
	return spk.suite
}

// Scalar returns the Scalar corresponding to this Private Key
func (spk *privateKey) Scalar() crypto.Scalar {
	return spk.sk
}

// Verify checks a signature over a message
func (pk *publicKey) Verify(data []byte, signature []byte, signer crypto.SingleSigner) error {
	if signer == nil {
		return crypto.ErrNilSingleSigner
	}

	return signer.Verify(pk.suite, pk.pk, data, signature)
}

// ToByteArray returns the byte array representation of the public key
func (pk *publicKey) ToByteArray() ([]byte, error) {
	return pk.pk.MarshalBinary()
}

// Suite returns the Suite (curve data) used for this private key
func (pk *publicKey) Suite() crypto.Suite {
	return pk.suite
}

// Point returns the Point corresponding to this Public Key
func (pk *publicKey) Point() crypto.Point {
	return pk.pk
}
